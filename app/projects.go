package main

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
)

type ProjectPane struct {
	*tview.Flex
	projects            []model.Project
	list                *tview.List
	newProject          *tview.InputField
	repo                repository.ProjectRepository
	activeProject       *model.Project
	projectListStarting int // The index in list where project names starts
}

func NewProjectPane(repo repository.ProjectRepository) *ProjectPane {
	pane := ProjectPane{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		list:       tview.NewList().ShowSecondaryText(false),
		newProject: makeLightTextInput("+[New Project]"),
		repo:       repo,
	}

	pane.newProject.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			pane.addNewProject()
		case tcell.KeyEsc:
			app.SetFocus(projectPane)
		}
	})

	pane.AddItem(pane.list, 0, 1, true).
		AddItem(pane.newProject, 1, 0, false)

	pane.SetBorder(true).SetTitle("[::u]P[::-]rojects")
	pane.loadListItems(false)

	return &pane
}

func (pane *ProjectPane) addNewProject() {
	name := pane.newProject.GetText()
	if len(name) < 3 {
		statusBar.showForSeconds("[red::]Project name should be at least 3 character", 5)
		return
	}

	project, err := pane.repo.Create(name, "")
	if err != nil {
		statusBar.showForSeconds("[red::]Failed to create Project:"+err.Error(), 5)
	} else {
		statusBar.showForSeconds(fmt.Sprintf("[yellow::]Project %s created. Press n to start adding new tasks.", name), 5)
		pane.projects = append(pane.projects, project)
		pane.addProjectToList(len(pane.projects)-1, true)
		pane.newProject.SetText("")
	}
}

func (pane *ProjectPane) addDynamicLists() {
	pane.addSection("Dynamic Lists")
	pane.list.AddItem("- Today", "", 0, yetToImplement("Today's Tasks"))
	pane.list.AddItem("- Upcoming", "", 0, yetToImplement("Upcoming Tasks"))
	pane.list.AddItem("- No Due Date", "", 0, yetToImplement("Unscheduled Tasks"))
}

func (pane *ProjectPane) addProjectList() {
	pane.addSection("Projects")
	pane.projectListStarting = pane.list.GetItemCount()

	var err error
	pane.projects, err = pane.repo.GetAll()
	if err != nil {
		statusBar.showForSeconds("Could not load Projects: "+err.Error(), 5)
		return
	}

	for i := range pane.projects {
		pane.addProjectToList(i, false)
	}

	pane.list.SetCurrentItem(pane.projectListStarting)
}

func (pane *ProjectPane) addProjectToList(i int, selectItem bool) {
	// To avoid overriding of loop variables - https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/
	pane.list.AddItem("- "+pane.projects[i].Title, "", 0, func(idx int) func() {
		return func() { pane.activateProject(idx) }
	}(i))

	if selectItem {
		pane.list.SetCurrentItem(-1)
		pane.activateProject(i)
	}
}

func (pane *ProjectPane) addSection(name string) {
	pane.list.AddItem("[::d]"+name, "", 0, nil)
	pane.list.AddItem("[::d]"+strings.Repeat(string(tcell.RuneS3), 25), "", 0, nil)
}

func (pane *ProjectPane) handleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'n':
		app.SetFocus(pane.newProject)
	}

	return event
}

func (pane *ProjectPane) activateProject(idx int) {
	pane.activeProject = &pane.projects[idx]
	taskPane.LoadProjectTasks(*pane.activeProject)

	removeThirdCol()
	projectDetailPane.SetProject(pane.activeProject)
	contents.AddItem(projectDetailPane, 25, 0, false)
	app.SetFocus(taskPane)
}

func (pane *ProjectPane) RemoveActivateProject() {
	if pane.activeProject != nil && pane.repo.Delete(pane.activeProject) == nil {

		for i := range taskPane.tasks {
			_ = taskRepo.Delete(&taskPane.tasks[i])
		}
		taskPane.ClearList()

		statusBar.showForSeconds("[lime]Removed Project: "+pane.activeProject.Title, 5)
		removeThirdCol()

		pane.loadListItems(true)
	}
}

func (pane *ProjectPane) loadListItems(focus bool) {
	pane.list.Clear()
	pane.addDynamicLists()
	pane.list.AddItem("", "", 0, nil)
	pane.addProjectList()

	if focus {
		app.SetFocus(pane)
	}
}

func (pane *ProjectPane) GetActiveProject() *model.Project {
	return pane.activeProject
}
