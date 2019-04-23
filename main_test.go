package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadLoginRelationsFromEnvNoLogins(t *testing.T) {
	os.Setenv("LOGIN_RELATION_1", "")
	os.Setenv("LOGIN_RELATION_2", "")
	var trelloSlackLoginRelations []loginRelation
	loadLoginRelationsFromEnv(&trelloSlackLoginRelations)

	assert.Equal(t, []loginRelation(nil), trelloSlackLoginRelations)
}

func TestLoadLoginRelationsFromEnv(t *testing.T) {
	os.Setenv("LOGIN_RELATION_1", "@test|slackId")

	var trelloSlackLoginRelations []loginRelation
	loadLoginRelationsFromEnv(&trelloSlackLoginRelations)

	assert.Equal(t, []loginRelation{{trello: "@test", slack: "slackId"}}, trelloSlackLoginRelations)
}

func TestLoadLoginRelationsFromEnvTwoLogins(t *testing.T) {
	os.Setenv("LOGIN_RELATION_1", "@test|slackId")
	os.Setenv("LOGIN_RELATION_2", "@test1|slackId1")

	var trelloSlackLoginRelations []loginRelation
	loadLoginRelationsFromEnv(&trelloSlackLoginRelations)

	assert.Equal(t, []loginRelation{{trello: "@test", slack: "slackId"}, {trello: "@test1", slack: "slackId1"}}, trelloSlackLoginRelations)
}
