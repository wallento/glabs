package config

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func (ac AccessLevel) String() string {
	if ac == 10 {
		return "guest"
	}
	if ac == 20 {
		return "reporter"
	}
	if ac == 30 {
		return "developer"
	}
	return "maintainer"
}

func GetAssignmentConfig(course, assignment string, onlyForStudentsOrGroups ...string) *AssignmentConfig {
	if !viper.IsSet(course) {
		log.Fatal().
			Str("course", course).
			Msg("configuration for course not found")
	}

	if !viper.IsSet(course + "." + assignment) {
		log.Fatal().
			Str("course", course).
			Str("assignment", assignment).
			Msg("configuration for assignment not found")
	}

	assignmentKey := course + "." + assignment
	per := per(assignmentKey)

	path := assignmentPath(course, assignment)
	url := viper.GetString("gitlab.host") + "/" + path

	containerRegistry := viper.GetBool(assignmentKey + ".containerRegistry")
	release := release(assignmentKey)
	if release != nil && release.DockerImages != nil {
		containerRegistry = true
	}

	assignmentConfig := &AssignmentConfig{
		Course:            course,
		Name:              assignment,
		Path:              path,
		URL:               url,
		Per:               per,
		Description:       description(assignmentKey),
		ContainerRegistry: containerRegistry,
		AccessLevel:       accessLevel(assignmentKey),
		Students:          students(per, course, assignment, onlyForStudentsOrGroups...),
		Groups:            groups(per, course, assignment, onlyForStudentsOrGroups...),
		Startercode:       startercode(assignmentKey),
		Clone:             clone(assignmentKey),
		Release:           release,
		Seeder:            seeder(assignmentKey),
	}

	return assignmentConfig
}

// Using email addresses instead of usernames/user-id's results in @ in the student's name.
// This is incompatible to the filesystem and gitlab so replacing the values is necessary.
func (cfg *AssignmentConfig) RepoSuffix(student *Student) string {
	if student.Email != nil {
		return strings.ReplaceAll(*student.Email, "@", "_at_")
	}
	if student.Id != nil {
		return fmt.Sprint(*student.Id)
	}
	if student.Username != nil {
		return *student.Username
	}

	return ""
}

func assignmentPath(course, assignment string) string {
	path := viper.GetString(course + ".coursepath")
	if semesterpath := viper.GetString(course + ".semesterpath"); len(semesterpath) > 0 {
		path += "/" + semesterpath
	}

	assignmentpath := path
	if group := viper.GetString(course + "." + assignment + ".assignmentpath"); len(group) > 0 {
		assignmentpath += "/" + group
	}

	return assignmentpath
}

func per(assignmentKey string) Per {
	if per := viper.GetString(assignmentKey + ".per"); per == "group" {
		return PerGroup
	}
	return PerStudent
}

func description(assignmentKey string) string {
	description := "generated by glabs"

	if desc := viper.GetString(assignmentKey + ".description"); desc != "" {
		description = desc
	}

	return description
}
