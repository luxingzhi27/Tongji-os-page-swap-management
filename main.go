package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/emirpasic/gods/queues/arrayqueue"
)

var PageManagementWay = 0 //0:FIFO 1:LRU
var PageManagementWayStr = binding.NewString()

const PageNum = 32
const PageSize = 10
const MemorySize = 4
const InstructionsListSize = 5
const InstructionsNum int = 320

var PageTable [][]int
var CurrentInstructionStr binding.String
var InstructionsList []binding.String
var MemoryButtons [][]*widget.Button
var LogicPageShowStr []binding.String
var SwapStatusShowStr binding.String
var MissingPagesNumStr binding.String
var missingPagesNum int
var InstructionSequence []int
var Speed binding.Float
var MemoryQueue *arrayqueue.Queue //FIFO
var MemoryFrequency []int         //LRU
var IsFull bool                   //LRU判断内存是否满
var Memory int                    //LRU内存占用情况
var LRUTime int                   //LRU时间戳

func dataInit() {
	PageTable = make([][]int, PageNum)
	for i := 0; i < PageNum; i++ {
		PageTable[i] = make([]int, PageSize)
		for j := 0; j < PageSize; j++ {
			PageTable[i][j] = i*PageSize + j
		}
	}
	missingPagesNum = 0
	MemoryQueue = arrayqueue.New()
	MemoryFrequency = make([]int, MemorySize)
	Memory = 0
	IsFull = false
	LRUTime = 0
	getSequence()
}

func reSet() {
	dataInit()
	for i := 0; i < MemorySize; i++ {
		for j := 0; j < PageSize; j++ {
			disHighlightButton(MemoryButtons[i][j])
			MemoryButtons[i][j].SetText("NULL")
		}
	}
	MissingPagesNumStr.Set("missing pages num: 0")
	missingPagesNum = 0
	SwapStatusShowStr.Set("swap status: ")
	CurrentInstructionStr.Set("current instruction: NULL")
	for i := 0; i < InstructionsListSize; i++ {
		InstructionsList[i].Set("NULL")
	}
}

func hilightButton(button *widget.Button) {
	button.Importance = widget.ButtonImportance(3)
	button.Refresh()
}

func disHighlightButton(button *widget.Button) {
	button.Importance = widget.ButtonImportance(2)
	button.Refresh()
}

func getSequence() {
	InstructionSequence = make([]int, InstructionsNum)
	rand.Seed(time.Now().UnixNano())
	var startIndex int
	// 随机选择起始执行指令序号
	startIndex = rand.Intn(InstructionsNum)

	currentIndex := startIndex

	for i := 0; i < InstructionsNum; i += 6 {
		InstructionSequence[i] = currentIndex
		if currentIndex == InstructionsNum-1 {
			currentIndex /= 2
		}
		currentIndex++
		InstructionSequence[i+1] = currentIndex
		if currentIndex == 1 {
			currentIndex = 0
		} else {
			currentIndex = rand.Intn(currentIndex - 1)
		}
		if i == InstructionsNum-2 {
			break
		}
		InstructionSequence[i+2] = currentIndex
		currentIndex++
		if currentIndex == InstructionsNum-1 {
			currentIndex /= 2
		}
		InstructionSequence[i+3] = currentIndex
		currentIndex = rand.Intn(InstructionsNum-currentIndex-1) + currentIndex + 1
		InstructionSequence[i+4] = currentIndex
		currentIndex++
		if currentIndex == InstructionsNum-1 {
			currentIndex /= 2
		}
		InstructionSequence[i+5] = currentIndex
		currentIndex = rand.Intn(currentIndex - 1)
	}
}

func getPageByInstruction(instruction int) int {
	return instruction/PageSize + 1
}

func startIterate() {
	for i := 0; i < InstructionsNum; i++ {
		CurrentInstructionStr.Set("current instruction: " + strconv.Itoa(InstructionSequence[i]))
		for j := 0; j < InstructionsListSize; j++ {
			index := i + j
			if index >= InstructionsNum {
				InstructionsList[j].Set("NULL")
			} else {
				InstructionsList[j].Set(strconv.Itoa(InstructionSequence[index]))
			}
		}
		speed, err := Speed.Get()
		if err != nil {
			panic(err)
		}
		timeInterval := float64(1500) / speed
		if PageManagementWay == 0 {
			clearImportance()
			FIFO(InstructionSequence[i])
		} else {
			clearImportance()
			LRU(InstructionSequence[i])
		}
		time.Sleep(time.Duration(timeInterval) * time.Millisecond)
	}
}

func checkInMemory(instruction int) bool {
	for i := 0; i < MemorySize; i++ {
		for j := 0; j < PageSize; j++ {
			if MemoryButtons[i][j].Text == strconv.Itoa(instruction) {
				hilightButton(MemoryButtons[i][j])
				return true
			}
		}
	}
	return false
}

func clearImportance() {
	for i := 0; i < MemorySize; i++ {
		for j := 0; j < PageSize; j++ {
			disHighlightButton(MemoryButtons[i][j])
		}
	}
}

func LRU(instruction int) {
	if checkInMemory(instruction) {
		SwapStatusShowStr.Set("instruction " + strconv.Itoa(instruction) + " is in memory")
		return
	} else {
		missingPagesNum++
		MissingPagesNumStr.Set("missing pages num: " + strconv.Itoa(missingPagesNum))
		if !IsFull {
			Memory++
			MemoryFrequency[Memory-1] = LRUTime
			if Memory == MemorySize {
				IsFull = true
			}
			replacePage(instruction, Memory)
		} else {
			pos := 1
			for i := 0; i < MemorySize; i++ {
				if MemoryFrequency[i] < MemoryFrequency[pos-1] {
					pos = i + 1
				}
			}
			MemoryFrequency[pos-1] = LRUTime
			replacePage(instruction, pos)
		}
	}
	LRUTime++ //时间戳加一
}

func FIFO(instruction int) {
	if checkInMemory(instruction) {
		SwapStatusShowStr.Set("instruction " + strconv.Itoa(instruction) + " is in memory")
		return
	} else {
		missingPagesNum++
		MissingPagesNumStr.Set("missing pages num: " + strconv.Itoa(missingPagesNum))
		lenth := MemoryQueue.Size()
		if lenth < MemorySize {
			MemoryQueue.Enqueue(lenth + 1)
			replacePage(instruction, lenth+1)
		} else {
			pos, _ := MemoryQueue.Dequeue()
			replacePage(instruction, pos.(int))
			MemoryQueue.Enqueue(pos)
		}
	}
}

// physicalPage: 1,2,3,4
func replacePage(instruction, physicalPage int) {
	logicPage := getPageByInstruction(instruction)
	SwapStatusShowStr.Set(fmt.Sprintf("swap logical page %d and physical page %d", logicPage, physicalPage))
	LogicPageShowStr[physicalPage-1].Set("logic page: " + strconv.Itoa(logicPage))
	for i := 0; i < PageSize; i++ {
		if PageTable[logicPage-1][i] == instruction {
			hilightButton(MemoryButtons[physicalPage-1][i])
		}
		MemoryButtons[physicalPage-1][i].SetText(strconv.Itoa(PageTable[logicPage-1][i]))
	}
}

func UI() {
	myApp := app.New()
	window := myApp.NewWindow("page swap management")
	lightTheme := theme.LightTheme()
	darkTheme := theme.DarkTheme()
	myApp.Settings().SetTheme(lightTheme)
	themeButton := widget.NewButton("Switch Theme", func() {
		if myApp.Settings().Theme() == lightTheme {
			myApp.Settings().SetTheme(darkTheme)
		} else {
			myApp.Settings().SetTheme(lightTheme)
		}
	})
	waySwitchButton := widget.NewButton("Switch Page Management Way", func() {
		if PageManagementWay == 0 {
			PageManagementWay = 1
			PageManagementWayStr.Set("LRU")
		} else {
			PageManagementWay = 0
			PageManagementWayStr.Set("FIFO")
		}
	})
	startButton := widget.NewButton("Start", func() {
		go startIterate()
	})
	startButton.Importance = widget.HighImportance
	resetButton := widget.NewButton("Reset", func() {
		reSet()
	})
	resetButton.Importance = widget.HighImportance
	buttonContainer := container.New(layout.NewHBoxLayout(), themeButton, layout.NewSpacer(), waySwitchButton, startButton, resetButton)

	memoryContainer := container.New(layout.NewHBoxLayout())
	LogicPageShowStr = make([]binding.String, MemorySize)
	MemoryButtons = make([][]*widget.Button, MemorySize)
	for i := 0; i < MemorySize; i++ {
		MemoryButtons[i] = make([]*widget.Button, PageSize)
	}
	for i := 0; i < MemorySize; i++ {
		singleMemoryContainer := container.New(layout.NewVBoxLayout())
		physicalPageShowlabel := widget.NewLabel("physical memory page" + strconv.Itoa(i+1))
		physicalPageShowlabel.Alignment = fyne.TextAlignCenter
		singleMemoryContainer.Add(physicalPageShowlabel)
		for j := 0; j < PageSize; j++ {
			MemoryButtons[i][j] = widget.NewButton("NULL", func() {})
			MemoryButtons[i][j].Alignment = widget.ButtonAlignCenter
			MemoryButtons[i][j].Importance = widget.ButtonImportance(2)
			singleMemoryContainer.Add(MemoryButtons[i][j])
		}
		LogicPageShowStr[i] = binding.NewString()
		LogicPageShowStr[i].Set("logic page: null")
		logicPageShowlabel := widget.NewLabelWithData(LogicPageShowStr[i])
		logicPageShowlabel.Alignment = fyne.TextAlignCenter
		singleMemoryContainer.Add(logicPageShowlabel)
		memoryContainer.Add(singleMemoryContainer)
	}
	instructionsListContainer := container.New(layout.NewVBoxLayout())
	instructionsListContainer.Add(layout.NewSpacer())
	instructionsListContainer.Add(widget.NewLabel("instructions list"))
	InstructionsList = make([]binding.String, InstructionsListSize)
	for i := 0; i < InstructionsListSize; i++ {
		InstructionsList[i] = binding.NewString()
		InstructionsList[i].Set("NULL")
		instructionLabel := widget.NewLabelWithData(InstructionsList[i])
		if i == 0 {
			instructionLabel.TextStyle.Bold = true
			instructionLabel.TextStyle.Italic = true
		}
		instructionLabel.Alignment = fyne.TextAlignCenter
		instructionsListContainer.Add(instructionLabel)
	}
	Speed = binding.NewFloat()
	speedSlider := widget.NewSliderWithData(1, 15, Speed)
	speedSlider.SetValue(8)
	instructionsListContainer.Add(widget.NewLabel("execute speed"))
	instructionsListContainer.Add(speedSlider)
	instructionsListContainer.Add(layout.NewSpacer())
	centerContainer := container.New(layout.NewHBoxLayout(), memoryContainer, layout.NewSpacer(), instructionsListContainer)
	PageManagementWayStr.Set("FIFO")
	wayStatus := widget.NewLabelWithData(PageManagementWayStr)
	wayStatus.TextStyle.Bold = true
	wayStatus.TextStyle.Italic = true
	CurrentInstructionStr = binding.NewString()
	CurrentInstructionStr.Set("current instruction: null")
	instructionStatus := widget.NewLabelWithData(CurrentInstructionStr)
	instructionStatus.TextStyle.Bold = true
	instructionStatus.TextStyle.Italic = true
	SwapStatusShowStr = binding.NewString()
	SwapStatusShowStr.Set("swap status: loaded 4 pages")
	swapStatus := widget.NewLabelWithData(SwapStatusShowStr)
	swapStatus.TextStyle.Bold = true
	swapStatus.TextStyle.Italic = true
	MissingPagesNumStr = binding.NewString()
	MissingPagesNumStr.Set("missing pages num: 0")
	missingPagesNumLabel := widget.NewLabelWithData(MissingPagesNumStr)
	missingPagesNumLabel.TextStyle.Bold = true
	missingPagesNumLabel.TextStyle.Italic = true
	statusBarContainer := container.New(layout.NewHBoxLayout(), wayStatus, layout.NewSpacer(), swapStatus, layout.NewSpacer(), missingPagesNumLabel, layout.NewSpacer(), instructionStatus)
	allContainer := container.New(layout.NewVBoxLayout(), buttonContainer, centerContainer, statusBarContainer)
	window.SetContent(allContainer)
	window.ShowAndRun()
}

func main() {

	dataInit()
	UI()
}
