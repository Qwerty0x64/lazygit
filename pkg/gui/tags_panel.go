package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
)

// list panel functions

func (gui *Gui) getSelectedTag() *commands.Tag {
	selectedLine := gui.State.Panels.Tags.SelectedLine
	if selectedLine == -1 || len(gui.State.Tags) == 0 {
		return nil
	}

	return gui.State.Tags[selectedLine]
}

func (gui *Gui) handleTagSelect() error {
	if gui.popupPanelFocused() {
		return nil
	}

	if gui.inDiffMode() {
		return gui.renderDiff()
	}

	var task updateTask
	tag := gui.getSelectedTag()
	if tag == nil {
		task = gui.createRenderStringTask("No tags")
	} else {
		cmd := gui.OSCommand.ExecutableFromString(
			gui.GitCommand.GetBranchGraphCmdStr(tag.Name),
		)
		task = gui.createRunCommandTask(cmd)
	}

	return gui.refreshMain(refreshMainOpts{
		main: &viewUpdateOpts{
			title: "Tag",
			task:  task,
		},
	})
}

func (gui *Gui) refreshTags() error {
	tags, err := gui.GitCommand.GetTags()
	if err != nil {
		return gui.surfaceError(err)
	}

	gui.State.Tags = tags

	return gui.postRefreshUpdate(gui.Contexts.Tags.Context)
}

func (gui *Gui) renderTagsContext() error {
	branchesView := gui.getBranchesView()

	gui.refreshSelectedLine(&gui.State.Panels.Tags.SelectedLine, len(gui.State.Tags))
	displayStrings := presentation.GetTagListDisplayStrings(gui.State.Tags, gui.State.Diff.Ref)
	gui.renderDisplayStrings(branchesView, displayStrings)

	return nil
}

func (gui *Gui) handleCheckoutTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.getSelectedTag()
	if tag == nil {
		return nil
	}
	if err := gui.handleCheckoutRef(tag.Name, handleCheckoutRefOptions{}); err != nil {
		return err
	}
	return gui.switchContext(gui.Contexts.Branches.Context)
}

func (gui *Gui) handleDeleteTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.getSelectedTag()
	if tag == nil {
		return nil
	}

	prompt := gui.Tr.TemplateLocalize(
		"DeleteTagPrompt",
		Teml{
			"tagName": tag.Name,
		},
	)

	return gui.ask(askOpts{
		returnToView:       v,
		returnFocusOnClose: true,
		title:              gui.Tr.SLocalize("DeleteTagTitle"),
		prompt:             prompt,
		handleConfirm: func() error {
			if err := gui.GitCommand.DeleteTag(tag.Name); err != nil {
				return gui.surfaceError(err)
			}
			return gui.refreshSidePanels(refreshOptions{mode: ASYNC, scope: []int{COMMITS, TAGS}})
		},
	})
}

func (gui *Gui) handlePushTag(g *gocui.Gui, v *gocui.View) error {
	tag := gui.getSelectedTag()
	if tag == nil {
		return nil
	}

	title := gui.Tr.TemplateLocalize(
		"PushTagTitle",
		Teml{
			"tagName": tag.Name,
		},
	)

	return gui.prompt(v, title, "origin", func(response string) error {
		if err := gui.GitCommand.PushTag(response, tag.Name); err != nil {
			return gui.surfaceError(err)
		}
		return nil
	})
}

func (gui *Gui) handleCreateTag(g *gocui.Gui, v *gocui.View) error {
	return gui.prompt(v, gui.Tr.SLocalize("CreateTagTitle"), "", func(tagName string) error {
		// leaving commit SHA blank so that we're just creating the tag for the current commit
		if err := gui.GitCommand.CreateLightweightTag(tagName, ""); err != nil {
			return gui.surfaceError(err)
		}
		return gui.refreshSidePanels(refreshOptions{scope: []int{COMMITS, TAGS}, then: func() {
			// find the index of the tag and set that as the currently selected line
			for i, tag := range gui.State.Tags {
				if tag.Name == tagName {
					gui.State.Panels.Tags.SelectedLine = i
					_ = gui.renderTagsContext()
					return
				}
			}
		},
		})
	})
}

func (gui *Gui) handleCreateResetToTagMenu(g *gocui.Gui, v *gocui.View) error {
	tag := gui.getSelectedTag()
	if tag == nil {
		return nil
	}

	return gui.createResetMenu(tag.Name)
}
