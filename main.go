package main

import (
	"1PanelHelper/notify"
	"encoding/json"
	"fmt"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

// RecoverGo 包装一个函数，使其在 goroutine 中安全运行
func RecoverGo(f func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("从panic中恢复: %v,推送错误消息", r)
				recordChannel <- recordInChannelStruct{
					recordItem: CronjobRecordItem{
						ID:        114514,
						StartTime: time.Now().Format(time.RFC3339),
						Message:   fmt.Sprintf("%v", r),
					},
					cronjobItem: CronjobItem{
						Name: "1PanelHelper自身错误",
					},
				}
			}
		}()
		f()
	}()
}

func getAllCronjobItems(p *PanelApiStruct) ([]CronjobItem, error) {
	log.Println("开始获取定时任务列表")
	list, err := p.GetCronjobList(SearchRequest{
		Page:     1,
		PageSize: 10,
		OrderBy:  "created_at",
		Order:    "null",
	})
	if err != nil {
		return nil, fmt.Errorf("获取定时任务列表失败: %w", err)
	}
	log.Printf("获取到部分定时任务列表，共 %d 条记录", list.Data.Total)

	fullList, err := p.GetCronjobList(SearchRequest{
		Page:     1,
		PageSize: list.Data.Total,
		OrderBy:  "created_at",
		Order:    "null",
	})
	if err != nil {
		return nil, fmt.Errorf("获取定时任务列表失败: %w", err)
	}
	log.Printf("获取到完整定时任务列表，共 %d 条记录", fullList.Data.Total)
	return fullList.Data.Items, nil
}

// notifyWorker 用于推送通知
func notifyWorker(n notify.Notify) {
	for item := range recordChannel {
		log.Printf("开始向通知通道推送任务 %d 的失败提醒", item.recordItem.ID)
		title := fmt.Sprintf("任务 「%s」出错了!", item.cronjobItem.Name)
		content := fmt.Sprintf("任务名: %s\n任务id: %s\n任务触发时间: %s\n错误信息: %s\n", item.cronjobItem.Name, strconv.Itoa(item.recordItem.ID), item.recordItem.StartTime, item.recordItem.Message)
		err := n.PushNotify(content, title)
		if err != nil {
			notifyErrorChannel <- fmt.Errorf("在推送消息时出错: %w", err)
		}
	}
}

type recordInChannelStruct struct {
	recordItem  CronjobRecordItem
	cronjobItem CronjobItem
}

var (
	jsonFileMutex      sync.Mutex
	recordChannel      = make(chan recordInChannelStruct, 200)
	notifyErrorChannel = make(chan error, 200)
)

// ProcessNewCronjobRecords 获取最新的记录并处理准备推送
func ProcessNewCronjobRecords(p *PanelApiStruct, dataFilePath string, cronjobId int, cronjobItem CronjobItem) error {
	log.Printf("开始处理任务 %d 的记录", cronjobId)
	records, err := p.GetCronjobRecords(CronjobRecordRequest{
		Page:      1,
		PageSize:  50,
		CronjobID: cronjobId,
		Days:      1,
	})
	if err != nil {
		return fmt.Errorf("从api获取任务记录失败: %w", err)
	}
	log.Printf("获取到任务 %d 的记录，共 %d 条", cronjobId, len(records.Data.Items))

	jsonFileMutex.Lock()
	defer jsonFileMutex.Unlock()

	var jsonData map[int]int
	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return fmt.Errorf("读取json文件时失败: %w", err)
	}

	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return fmt.Errorf("反序列化json文件失败: %w", err)
	}

	newestIdInJsonFile := jsonData[cronjobId]

	var newRecordList []CronjobRecordItem
	for _, record := range records.Data.Items {
		if record.ID != newestIdInJsonFile {
			newRecordList = append(newRecordList, record)
			continue
		}
		break
	}
	if len(records.Data.Items) == 0 {
		log.Printf("任务 %d 一天内还没有运行过,跳过", cronjobId)
		return nil
	}
	newestIdInJsonFile = records.Data.Items[0].ID
	// 更新JSON文件
	jsonData[cronjobId] = newestIdInJsonFile
	updatedData, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("序列化更新后的数据失败: %w", err)
	}

	err = os.WriteFile(dataFilePath, updatedData, 0644)
	if err != nil {
		return fmt.Errorf("写入更新后的数据到文件失败: %w", err)
	}
	log.Printf("更新任务 %d 的记录到JSON文件", cronjobId)

	for _, item := range newRecordList {
		if item.Status != "Success" && item.Status != "Waiting" {
			log.Printf("任务 %d 的运行状态非 Success/Waiting,有问题!", item.ID)
			recordChannel <- recordInChannelStruct{recordItem: item, cronjobItem: cronjobItem}
		}
	}

	return nil
}

func main() {
	p := PanelApiStruct{
		BaseUrl: os.Getenv("PANEL_BASE_URL"),
	}
	err := p.LoginWithout2FA(os.Getenv("PANEL_USERNAME"), os.Getenv("PANEL_PASSWORD"), os.Getenv("PANEL_ENTRANCE_CODE"))
	if err != nil {
		log.Fatalf("登录失败: %v", err)
	}
	log.Println("登录成功")

	dataFilePath := "cronjob_records.json"
	checkInterval := 1 * time.Hour

	// 初始化通知系统
	n := notify.GotifyStruct{
		ServerRootUrl: os.Getenv("GOTIFY_BASE_URL"),
		Token:         os.Getenv("GOTIFY_APP_TOKEN"),
	}
	n.Priority, _ = strconv.Atoi(os.Getenv("GOTIFY_PRIORITY"))
	log.Println("通知系统初始化成功")

	// 启动notify worker
	RecoverGo(func() {
		notifyWorker(n)
	})
	log.Println("notify worker启动成功")

	// 启动错误日志监控
	RecoverGo(func() {
		monitorNotifyErrors()
	})
	log.Println("错误日志监控启动成功")

	for {
		cronjobItems, err := getAllCronjobItems(&p)
		if err != nil {
			log.Printf("获取定时任务列表失败: %v", err)
			time.Sleep(checkInterval)
			continue
		}
		log.Printf("获取到定时任务列表，共 %d 条记录", len(cronjobItems))

		var wg sync.WaitGroup
		for _, item := range cronjobItems {
			wg.Add(1)
			RecoverGo(func() {
				defer wg.Done()
				err := ProcessNewCronjobRecords(&p, dataFilePath, item.ID, item)
				if err != nil {
					log.Printf("更新任务 %d 的记录失败: %v", item.ID, err)
				}
			})
		}

		wg.Wait()
		log.Println("所有任务检查完成，等待下一次检查")
		time.Sleep(checkInterval)
	}
}

// monitorNotifyErrors 监控通知错误并记录日志
func monitorNotifyErrors() {
	for err := range notifyErrorChannel {
		log.Printf("通知错误: %v", err)
	}
}

// initializeJSONFile 确保JSON文件存在并包含有效的数据
func initializeJSONFile(filePath string) error {
	log.Printf("检查JSON文件 %s 是否存在", filePath)
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Printf("JSON文件 %s 不存在，创建新文件", filePath)
		emptyData := map[int]int{}
		jsonData, err := json.Marshal(emptyData)
		if err != nil {
			return fmt.Errorf("创建空JSON数据失败: %w", err)
		}
		return os.WriteFile(filePath, jsonData, os.ModePerm)
	}
	return err
}

func init() {
	// 确保JSON文件存在
	err := initializeJSONFile("cronjob_records.json")
	if err != nil {
		log.Fatalf("初始化JSON文件失败: %v", err)
	}
	log.Println("JSON文件初始化成功")
}
