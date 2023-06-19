package zaplog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/natefinch/lumberjack"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level zapcore.Level

// log層級
const (
	DebugLevel Level = Level(zap.DebugLevel)
	InfoLevel  Level = Level(zap.InfoLevel)
	ErrorLevel Level = Level(zap.ErrorLevel)
)

// var會比init早執行,但zaplog這裡不是用mconfig,所以先給預設值
var (
	maxlogsize      int
	maxbackup       int
	maxage          int
	svcname         string
	logFilePath     string
	defaultLogLevel string
)

var (
	logger      *zap.SugaredLogger //log實例
	atomicLevel = zap.NewAtomicLevel()

	levelMap = map[string]zapcore.Level{
		"debug": zapcore.DebugLevel,
		"info":  zapcore.InfoLevel,
		"error": zapcore.ErrorLevel,
	}
	elsService = NewElasticService()
	zviper     *viper.Viper
)

const (
	configFileName       = "config"
	viperReadFileError   = "viper read config file error:%v"
	viperReadConfigError = "viper read config error ,configPath:%s ,data:%v"
	configChangeMessage  = "config file changed ,data:%s"
)

// 初始化viper,因為zaplog為最底層,跟mconfig又不同包,避免循環參照只能拉到同一層或是獨立viper
func InitZaplog() {
	zviper = viper.New()
	zviper.AddConfigPath("./")
	zviper.SetConfigName(configFileName)
	err := zviper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf(viperReadFileError, err))
	}
	maxlogsize = zviper.GetInt("log.maxlogsize")              //每50MB切割log檔
	maxbackup = zviper.GetInt("log.maxbackup")                //達300個切割開始取代舊檔
	maxage = zviper.GetInt("log.maxage")                      //log保存最大時間(days)
	svcname = zviper.GetString("log.svcname")                 //log檔名
	logFilePath = zviper.GetString("log.logFilePath")         //log檔案路徑
	defaultLogLevel = zviper.GetString("log.defaultLogLevel") //預設log level
	filePath := getFilePath()
	level := getLoggerLevel(defaultLogLevel)
	log := NewLogger(filePath, level, maxlogsize, maxbackup, maxage, true, svcname)
	logger = log.Sugar()
	logger.Sync()
	initElastic()
}

// 取logger層級,預設info
func getLoggerLevel(lvl string) zapcore.Level {
	if level, ok := levelMap[lvl]; ok {
		return level
	}
	return zapcore.InfoLevel
}

// 用於動態調整logger層級
// *TODO 提供一個API便於程式執行時切換logger層級
func SetLoggerLevel(lvl string) {
	filePath := getFilePath()
	level := getLoggerLevel(lvl)
	log := NewLogger(filePath, level, maxlogsize, maxbackup, maxage, true, svcname)
	logger = log.Sugar()
	logger.Sync()
}

// 建立zap的實例,添加caller顯示調用log的代碼位置
func NewLogger(filePath string, level zapcore.Level, maxSize int, maxBackups int, maxAge int, compress bool, serviceName string) *zap.Logger {
	core := newCore(filePath, level, maxSize, maxBackups, maxAge, compress)
	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
}

// 設置zapcore,添加輸出到os.Stdout跟lumberjack log切割
func newCore(filePath string, level zapcore.Level, maxSize int, maxBackups int, maxAge int, compress bool) zapcore.Core {
	hook := lumberjack.Logger{
		Filename:   filePath,   // 日志文件路径
		MaxSize:    maxSize,    // 每个日志文件保存的最大尺寸 单位：M
		MaxBackups: maxBackups, // 日志文件最多保存多少个备份
		MaxAge:     maxAge,     // 文件最多保存多少天
		Compress:   compress,   // 是否压缩
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,    // 全大寫
		EncodeTime:     zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式,2006-01-02T15:04:05.000Z0700
		EncodeDuration: zapcore.SecondsDurationEncoder, //顯示浮點樹的秒
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短調用路徑
		// EncodeCaller:   zapcore.FullCallerEncoder,    // 長調用路徑
		EncodeName: zapcore.FullNameEncoder,
	}
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig), // 编码器配置
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(elsService), zapcore.AddSync(&hook)), // 打印到控制台和文件
		zap.NewAtomicLevelAt(level), // 日志级别
	)
}

// 取API執行的當前目錄
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		logger.Error(err)
	}
	return dir
}

// 取設定的log檔案路徑
func getFilePath() string {
	logfile := fmt.Sprintf(logFilePath, getCurrentDirectory(), svcname)
	return logfile
}

// 輸出debug log,單一參數轉json
func Debug(arg interface{}) {
	data, err := json.Marshal(arg)
	if err != nil {
		logger.Debug(arg)
		return
	}
	logger.Debug(string(data))
}

// 輸出debug log,類似fmt.Sprintf
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// 輸出debug log,zap的keyvalue風格,配合一些log portal查詢較明確不須全文本查詢
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// 輸出info log,單一參數轉json
func Info(arg interface{}) {
	data, err := json.Marshal(arg)
	if err != nil {
		logger.Info(arg)
		return
	}
	logger.Info(string(data))
}

// 輸出info log,類似fmt.Sprintf
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// 輸出info log,zap的keyvalue風格,配合一些log portal查詢較明確不須全文本查詢
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// 輸出error log,單一參數轉json
func Error(arg interface{}) {
	data, err := json.Marshal(arg)
	if err != nil {
		logger.Error(arg)
		return
	}
	logger.Error(string(data))
}

// 輸出error log,類似fmt.Sprintf
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// 輸出error log,zap的keyvalue風格,配合一些log portal查詢較明確不須全文本查詢
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}
