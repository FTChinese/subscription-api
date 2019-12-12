package query

import "fmt"

// Chooses which DB to use depending on running environment.
func getMembershipDB(sandbox bool) string {
	if sandbox {
		return "sandbox"
	}

	return "premium"
}

func BuildSelectMembership(sandbox bool, lock bool) string {

	suffix := ""
	if lock {
		suffix = "FOR UPDATE"
	}

	return fmt.Sprintf(
		selectMembership,
		getMembershipDB(sandbox),
		"vip_id",
		suffix)
}

func BuildSelectAppleMembership(sandbox bool) string {
	return fmt.Sprintf(
		selectMembership,
		getMembershipDB(sandbox),
		"apple_subscription_id",
		"FOR UPDATE")
}

func BuildInsertMembership(sandbox bool) string {
	return fmt.Sprintf(
		insertMembership,
		getMembershipDB(sandbox),
	)
}

func BuildUpdateMembership(sandbox bool) string {
	return fmt.Sprintf(
		updateMembership,
		getMembershipDB(sandbox))
}

func BuildDeleteMembership(sandbox bool) string {
	return fmt.Sprintf(
		deleteFtcMembership,
		getMembershipDB(sandbox))
}

func BuildUpdateMembershipID(sandbox bool) string {
	return fmt.Sprintf(
		updateMembershipID,
		getMembershipDB(sandbox))
}

func BuildInsertMemberSnapshot(sandbox bool) string {
	return fmt.Sprintf(
		insertMemberSnapshot,
		getMembershipDB(sandbox))
}

func BuildInsertClientApp(sandbox bool) string {
	return fmt.Sprintf(
		insertClientApp,
		getMembershipDB(sandbox))
}

func BuildInsertOrder(sandbox bool) string {
	return fmt.Sprintf(
		insertOrder,
		getMembershipDB(sandbox))
}

func BuildSelectOrder(sandbox bool) string {
	return fmt.Sprintf(
		selectOrder,
		getMembershipDB(sandbox))
}

func BuildConfirmOrder(sandbox bool) string {
	return fmt.Sprintf(
		updateConfirmedOrder,
		getMembershipDB(sandbox))
}

func BuildActivateGiftCard(sandbox bool) string {
	return fmt.Sprintf(
		activateGiftCard,
		getMembershipDB(sandbox))
}

func BuildInsertProration(sandbox bool) string {
	return fmt.Sprintf(
		insertProration,
		getMembershipDB(sandbox))
}

func BuildSelectBalanceSource(sandbox bool) string {
	return fmt.Sprintf(
		selectBalanceSource,
		getMembershipDB(sandbox),
		getMembershipDB(sandbox))
}

func BuildProrationUsed(sandbox bool) string {
	return fmt.Sprintf(
		prorationUsed,
		getMembershipDB(sandbox))
}

func BuildSelectProration(sandbox bool) string {
	return fmt.Sprintf(
		selectProration,
		getMembershipDB(sandbox))
}

func BuildInsertUpgradeBalance(sandbox bool) string {
	return fmt.Sprintf(
		insertUpgradeSchema,
		getMembershipDB(sandbox))
}

func BuildSelectUpgradePlan(sandbox bool) string {
	return fmt.Sprintf(
		selectUpgradeSchema,
		getMembershipDB(sandbox))
}
