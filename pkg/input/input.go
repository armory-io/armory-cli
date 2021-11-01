package input

import (
	"errors"
	"fmt"
	"github.com/manifoldco/promptui"
	"regexp"
)

type PromptMsg struct {
	Text    string
	ErrorMsg string
}


func PromptConfirmInput(pm PromptMsg) (bool, error) {
	val := func(input string) error {
		if len(input) > 3 {
			return errors.New(pm.ErrorMsg)
		}
		return nil
	}
	faintText := promptui.Styler(promptui.FGFaint)
	boldText := promptui.Styler(promptui.FGBold)
	greenText := promptui.Styler(promptui.FGGreen)
	redText := promptui.Styler(promptui.FGRed)
	defaultLabel := fmt.Sprintf("%s %s ", boldText(pm.Text), faintText("[y/N]"))

	template := promptui.PromptTemplates{
		Prompt:  defaultLabel,
		Valid:   defaultLabel,
		Invalid: fmt.Sprintf("%s %s ", redText(pm.Text), faintText("[y/N]")),
		Success: greenText(pm.Text) + " ",
	}

	prompt := promptui.Prompt{
		Validate:  val,
		Templates: &template,
	}
	result, err := prompt.Run()
	if err != nil {
		return false, err
	}
	return isConfirm(result), nil
}

var confirmExp = regexp.MustCompile("(?i)y(?:es)?|1")

func isConfirm(confirm string) bool {
	return confirmExp.MatchString(confirm)
}