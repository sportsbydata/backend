package scouting

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_modesSubmodesConflicts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		First     modeSubmode
		Second    modeSubmode
		Available []modeSubmode
	}{
		{
			First: modeSubmode{
				mode:    ModeAttack,
				submode: SubmodeOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeDefence,
				submode: SubmodeOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttack,
				submode: SubmodeNotOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeDefence,
				submode: SubmodeNotOurRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeDefence,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeNotOurRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAnyRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodeAllRules,
				},
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttackDefence,
				submode: SubmodeAnyRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		{
			First: modeSubmode{
				mode:    ModeAttackDefence,
				submode: SubmodeAllRules,
			},
			Available: []modeSubmode{
				{
					mode:    ModeAttack,
					submode: SubmodePlays,
				},
				{
					mode:    ModeDefence,
					submode: SubmodePlays,
				},
				{
					mode:    ModeAttackDefence,
					submode: SubmodePlays,
				},
			},
		},
		// TODO: add cases for plays
	}

	buildUnavailable := func(skip []modeSubmode) []modeSubmode {
		var mm []modeSubmode

		for _, m := range []Mode{ModeAttack, ModeDefence, ModeAttackDefence} {
			for _, sm := range []Submode{SubmodeOurRules, SubmodeNotOurRules, SubmodeAnyRules, SubmodeAllRules, SubmodePlays} {
				if !modeSubmodeValid(m, sm) {
					continue
				}

				var ignore bool

				for _, sk := range skip {
					if sk.mode == m && sk.submode == sm {
						ignore = true

						break
					}
				}

				if ignore {
					continue
				}

				mm = append(mm, modeSubmode{m, sm})
			}
		}

		return mm
	}

	for _, tc := range cases {
		baseName := fmt.Sprintf("(%s+%s)", tc.First.mode, tc.First.submode)

		for _, av := range tc.Available {
			fullName := fmt.Sprintf("%s with (%s+%s)", baseName, av.mode, av.submode)

			t.Run(fullName, func(t *testing.T) {
				assert.False(t, modesSubmodesConflicts(tc.First, av), "should not conflict")
			})
		}

		for _, sm := range buildUnavailable(tc.Available) {
			fullName := fmt.Sprintf("%s with (%s+%s)", baseName, sm.mode, sm.submode)

			t.Run(fullName, func(t *testing.T) {
				assert.True(t, modesSubmodesConflicts(tc.First, sm), "should conflict")
			})
		}
	}
}
