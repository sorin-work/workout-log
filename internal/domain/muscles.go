package domain

// Muscle groups used for weekly volume. Every exercise has one primary group
// and any number of secondary ones.
var MuscleGroups = []string{
	"chest", "back", "shoulders", "biceps", "triceps", "forearms",
	"quads", "hamstrings", "glutes", "calves", "core",
}

// Equipment types an exercise can use.
var EquipmentTypes = []string{
	"barbell", "dumbbell", "machine", "cable", "bodyweight", "kettlebell", "band", "other",
}

// IsMuscleGroup reports whether s is a known muscle group.
func IsMuscleGroup(s string) bool { return contains(MuscleGroups, s) }

// IsEquipment reports whether s is a known equipment type.
func IsEquipment(s string) bool { return contains(EquipmentTypes, s) }

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
