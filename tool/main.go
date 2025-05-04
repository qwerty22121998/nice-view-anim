package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Map[T any](arr []T, fn func(T) T) []T {
	res := make([]T, len(arr))
	for i, v := range arr {
		res[i] = fn(v)
	}
	return res
}

type extractor struct {
	originFile    string
	baseFileName  string
	mergeFileName string
	dstDir        string
	totalFrames   int
	frameDelay    int
	fileNames     []string
}

func (e *extractor) exec() error {
	e.baseFileName = strings.TrimSuffix(e.originFile, filepath.Ext(e.originFile))
	e.mergeFileName = fmt.Sprintf("%v_art.c", e.baseFileName)
	if err := os.MkdirAll(e.dstDir, os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(e.dstDir, "images"), os.ModePerm); err != nil {
		return err
	}
	return e.extract()
}

func (e *extractor) extractFrame(frame *image.Paletted, index int) error {
	fileName := fmt.Sprintf("%s%d.png", e.baseFileName, index)
	dstFileName := filepath.Join(e.dstDir, "images", fileName)
	f, err := os.Create(dstFileName)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := png.Encode(f, frame); err != nil {
		return err
	}
	e.fileNames = append(e.fileNames, fileName)
	return nil
}

func (e *extractor) lvglConvert() error {
	args := []string{
		"run", "--rm", "-u", "1000:1000", "-v", fmt.Sprintf("%v:/usr/src/proj", e.dstDir), "lv_img_conv",
		"-f", "-d", "-c", "CF_INDEXED_1_BIT",
	}
	args = append(args, Map(e.fileNames, func(s string) string {
		return filepath.Join("images", s)
	})...)
	if err := exec.Command("docker", args...).Run(); err != nil {
		return err
	}
	//for _, fileName := range e.fileNames {
	//	if err := os.Remove(filepath.Join(e.dstDir, fileName)); err != nil {
	//		return err
	//	}
	//}
	return nil
}

func (e *extractor) merge() error {
	dstFile := fmt.Sprintf("%v_art.c", e.baseFileName)
	dstFile = filepath.Join(e.dstDir, dstFile)
	f, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(`#include <lvgl.h>
#include "art.h"

#ifndef LV_ATTRIBUTE_MEM_ALIGN
#define LV_ATTRIBUTE_MEM_ALIGN
#endif
`); err != nil {
		return err
	}

	for i := 0; i < e.totalFrames; i++ {
		srcFileName := fmt.Sprintf("%v%d.c", e.baseFileName, i)
		srcFileName = filepath.Join(e.dstDir, srcFileName)
		content, err := os.ReadFile(srcFileName)
		lines := strings.Split(string(content), "\n")
		lines = lines[19:]
		if err != nil {
			return err
		}
		idx1 := lines[5]
		idx2 := lines[6]
		lines = append(lines[:5], append([]string{"#if CONFIG_NICE_VIEW_WIDGET_INVERTED", idx1, idx2, "#else", idx2, idx1, "#endif"}, lines[7:]...)...)
		if _, err := f.WriteString(strings.Join(lines, "\n")); err != nil {
			return err
		}
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
		if err := os.Remove(srcFileName); err != nil {
			return err
		}
	}
	if _, err := f.WriteString(fmt.Sprintf("const struct nice_view_anim anim_%v = {\n", e.baseFileName)); err != nil {
		return err
	}

	if _, err := f.WriteString(fmt.Sprintf(".len = %d,\n.duration = %d,\n.imgs = (const lv_img_dsc_t *[]){\n", e.totalFrames, e.totalFrames*e.frameDelay)); err != nil {
		return err
	}

	for i := 0; i < e.totalFrames; i++ {
		if _, err := f.WriteString(fmt.Sprintf("&%v%d,\n", e.baseFileName, i)); err != nil {
			return err
		}
	}

	if _, err := f.WriteString("},\n};\n"); err != nil {
		return err
	}
	return nil
}

func (e *extractor) extract() error {
	f, err := os.Open(e.originFile)
	if err != nil {
		return err
	}
	defer f.Close()
	img, err := gif.DecodeAll(f)
	if err != nil {
		return err
	}
	e.totalFrames = len(img.Image)
	e.frameDelay = img.Delay[0] * 10
	for i, frame := range img.Image {
		if err := e.extractFrame(frame, i); err != nil {
			return err
		}
	}

	if err := e.lvglConvert(); err != nil {
		return err
	}

	if err := e.merge(); err != nil {
		return err
	}

	return nil
}

func main() {
	e := &extractor{
		originFile: os.Args[1],
		dstDir:     "./build",
	}
	if err := e.exec(); err != nil {
		panic(err)
	}
}
