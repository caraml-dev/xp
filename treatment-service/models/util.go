package models

func ContainsProjectId(slice []ProjectId, item ProjectId) bool {
	if len(slice) == 0 {
		return true
	}

	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
