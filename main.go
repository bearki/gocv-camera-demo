package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gocv.io/x/gocv"
)

// CameraObject 相机对象
type CameraObject struct {
	Buffer       []byte       // 文件流
	BufferMutexx sync.RWMutex // 文件流读写锁(不加锁会出现图片闪屏问题)
}

// caOb 相机对象实例化
var caOb CameraObject

func main() {
	// 初始化相机设备
	go initCamerDaemon()
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/go", func(ctx *gin.Context) {
		// ctx.Status(200)
		// ctx.Writer.Write(GetBuffer())
		// // ctx.Header("Content-Type", "image/jpeg")
		ctx.Data(200, "image/jpeg", GetBuffer())
		// ctx.Abort()
	})
	r.Run(":8000")
}

// initCamerDaemon 相机守护,相机启动从这里开始
func initCamerDaemon() {
	// 死循环，一旦相机初始化函数接收就立即重启
	for {
		startCamersDevice()
		fmt.Println("相机设备意外停止了，正在重启···")
	}
}

// startCamersDevice 启动相机设备
func startCamersDevice() {
	// 捕获异常
	defer func() {
		if errs := recover(); errs != nil {
			fmt.Println(errs)
		}
	}()
	// 默认设备0，摄像头枚举范围（0-4），最多枚举到第5个设备
	videoDevice := 0
	// 打个标签，一会从这里重新开始
RETRY:
	// 开始尝试打开摄像头
	capDevice, err := gocv.VideoCaptureDevice(videoDevice)
	if err != nil {
		fmt.Printf("打开设[id: %d]备失败\n", videoDevice)
		// 设备打开失败，开始下一个相机
		videoDevice++
		// 判断设备枚举是否已超过第5个
		if videoDevice == 5 {
			// 重新回归0
			videoDevice = 0
		}
		// 重新回去打开
		goto RETRY
	}
	// 摄像机启动成功
	fmt.Println("相机启动完成")
	// 初始化一个mat对象
	mat := gocv.NewMat()
	// 死循环读取流
	for {
		// 暂停16毫秒
		time.Sleep(time.Millisecond * 16)
		// 读取流
		ok := capDevice.Read(&mat)
		if ok {
			// 开启写锁
			caOb.BufferMutexx.Lock()
			// mat对象转buffer
			buf, err := gocv.IMEncode(".jpg", mat)
			if err != nil {
				fmt.Println("mat对象转buffer失败")
				// 跳过
				continue
			}
			// 清理一遍老图
			caOb.Buffer = nil
			caOb.Buffer = make([]byte, len(buf.GetBytes()))
			// 赋值buffer
			copy(caOb.Buffer, buf.GetBytes())
			// 关闭写锁
			caOb.BufferMutexx.Unlock()
			// 关闭buffer
			buf.Close()
		} else {
			// 记录下日志
			fmt.Println("读取文件流失败")
			// 重新去打开摄像头
			goto RETRY
		}
	}
}

// GetBuffer 获取摄像头文件流
// @return []byte 图像文件流
func GetBuffer() []byte {
	// 开启读锁
	caOb.BufferMutexx.RLock()
	// 赋值文件流
	buf := make([]byte, len(caOb.Buffer))
	copy(buf, caOb.Buffer)
	// 关闭读锁
	caOb.BufferMutexx.RUnlock()
	// 返回文件流
	return buf
}
