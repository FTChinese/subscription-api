package subrepo

import "fmt"

// Chooses which DB to use depending on running environment.
func getMembershipDB(sandbox bool) string {
	if sandbox {
		return "sandbox"
	}

	return "premium"
}

func buildSelectMembership(sandbox bool, lock bool) string {

	suffix := ""
	if lock {
		suffix = "FOR UPDATE"
	}

	return fmt.Sprintf(
		stmtSelectMembership,
		getMembershipDB(sandbox),
		suffix)
}

func buildInsertMembership(sandbox bool) string {
	return fmt.Sprintf(
		stmtInsertMembership,
		getMembershipDB(sandbox),
	)
}

func buildUpdateMembership(sandbox bool) string {
	return fmt.Sprintf(
		stmtUpdateMembership,
		getMembershipDB(sandbox))
}

func buildUpdateMembershipID(sandbox bool) string {
	return fmt.Sprintf(
		stmtUpdateMembershipID,
		getMembershipDB(sandbox))
}

func buildInsertMemberSnapshot(sandbox bool) string {
	return fmt.Sprintf(
		stmtInsertMemberSnapshot,
		getMembershipDB(sandbox))
}
