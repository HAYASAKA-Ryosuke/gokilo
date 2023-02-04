package autoCompletion

import (
	"gokilo/lsp"
	"log"
)

type CompletionInfo struct {
	icon string
	text string
}

type AutoCompletionList struct {
	list          []string
	selectedIndex int
}

type AutoCompletion struct {
	autoCompletionEnable bool
	autoCompletionList   AutoCompletionList
}

func NewAutoCompletion() *AutoCompletion {
	return &AutoCompletion{autoCompletionEnable: false, autoCompletionList: AutoCompletionList{list: []string{}, selectedIndex: 0}}
}

func (a *AutoCompletion) SetEnabled(enable bool) {
	a.autoCompletionEnable = enable
	a.autoCompletionList.selectedIndex = 0
}

func (a *AutoCompletion) IsEnabled() bool {
	return a.autoCompletionEnable
}

func (a *AutoCompletion) GetCompletions(maxLength int) ([]string, int, int, int) {
	if !a.autoCompletionEnable {
		return []string{}, 0, 0, 0
	}
	if len(a.autoCompletionList.list) > 0 {
		if a.autoCompletionList.selectedIndex >= maxLength {
			return a.autoCompletionList.list[a.autoCompletionList.selectedIndex-maxLength+1 : a.autoCompletionList.selectedIndex+1], maxLength - 1, a.autoCompletionList.selectedIndex, len(a.autoCompletionList.list)
		} else {
			return a.autoCompletionList.list, a.autoCompletionList.selectedIndex, a.autoCompletionList.selectedIndex, len(a.autoCompletionList.list)
		}
	}
	return []string{}, 0, 0, 0
}

func (a *AutoCompletion) UpdateIndex(additional int) int {
	if !a.autoCompletionEnable {
		return a.autoCompletionList.selectedIndex
	}
	if (a.autoCompletionList.selectedIndex+additional >= 0) && (a.autoCompletionList.selectedIndex+additional < len(a.autoCompletionList.list)) {
		a.autoCompletionList.selectedIndex += additional
	}
	return a.autoCompletionList.selectedIndex
}

func (a *AutoCompletion) UpdateAutoCompletion(path string, lsp *lsp.Lsp, row int, column int) {
	if !a.autoCompletionEnable {
		return
	}
	completionList := lsp.Completion(path, uint32(row), uint32(column))
	log.Println("completion!")
	if len(completionList.Items) > 0 {
		completionItems := []string{}
		for _, item := range completionList.Items {
			completionItems = append(completionItems, item.Label)
		}
		log.Println(completionItems)
		a.autoCompletionList = AutoCompletionList{list: completionItems, selectedIndex: 0}
	}
}
