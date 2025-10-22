package api

import (
	"sync"

	"github.com/go-ego/gse"
)

var (
	segmenter     gse.Segmenter
	segmenterOnce sync.Once
)

// InitSegmenter 初始化分词器（全局单例）
func InitSegmenter() {
	segmenterOnce.Do(func() {
		segmenter.LoadDict() // 加载默认词典
	})
}

// GetSegmenter 获取分词器实例
func GetSegmenter() *gse.Segmenter {
	InitSegmenter()
	return &segmenter
}
