package ui

import (
	"errors"
	"strings"

	"github.com/charmbracelet/huh"
)

func Select(options []string, title string) (string, error) {
	var selectedOption string
	var optionsList []huh.Option[string]

	for _, option := range options {
		optionsList = append(optionsList, huh.NewOption(option, option))
	}

	theme := huh.ThemeCatppuccin()
	form := huh.NewSelect[string]().
		Title(title).
		Options(
			optionsList...,
		).
		Value(&selectedOption).
		WithTheme(theme)

	err := form.Run()

	return selectedOption, err
}

func SingleQuestion(question, placeholder string) (string, error) {
	theme := huh.ThemeCatppuccin()
	var answer string

	form := huh.NewInput().
		Inline(true).
		Title(question).
		Placeholder(placeholder).
		Validate(func(value string) error {
			if value == "" {
				return errors.New("value cannot be empty")
			}
			return nil
		}).
		Value(&answer).
		WithTheme(theme)

	err := form.Run()

	return answer, err
}

func GetGitDetails() (string, string, error) {
	theme := huh.ThemeCatppuccin()
	var username string
	var email string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Inline(true).
				Title("Username").
				Placeholder("Username").
				Validate(func(value string) error {
					if value == "" {
						return errors.New("username is required")
					}
					return nil
				}).
				Value(&username),
			huh.NewInput().Inline(true).
				Title("Email").
				Placeholder("Email").
				Validate(func(value string) error {
					if value == "" {
						return errors.New("email is required")
					}
					return nil
				}).
				Value(&email),
		),
	)

	err := form.WithTheme(theme).Run()
	return username, email, err
}

type PRDetails struct {
	Title     string
	Body      string
	Labels    []string
	Reviewers []string
	Draft     bool
}

func CreatePR(branchTitle, prTemplate string) (*PRDetails, error) {
	theme := huh.ThemeCatppuccin()

	var title, body string
	var tmpLabels, tmpReviewers string
	var draft bool

	if len(prTemplate) > 0 {
		body = strings.TrimSpace(prTemplate)
	}

	var labels, reviewers []string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Inline(true).
				Title("Enter a title:").
				Placeholder(branchTitle).
				Value(&title),
			huh.NewText().
				Title("Enter a body:").
				Value(&body),
			huh.NewInput().Inline(true).
				Title("Enter labels (comma separated):").
				Value(&tmpLabels),
			huh.NewInput().Inline(true).
				Title("Enter reviewers (comma separated):").
				Value(&tmpReviewers),
			huh.NewConfirm().Title("Is this a draft PR?").Value(&draft),
		),
	)

	labels = strings.Split(tmpLabels, ",")
	reviewers = strings.Split(tmpReviewers, ",")

	err := form.WithTheme(theme).Run()

	if title == "" {
		title = branchTitle
	}

	return &PRDetails{
		title,
		body,
		labels,
		reviewers,
		draft,
	}, err
}
