package upgrade

// UpgradeTask interfaces gives a set of methods to be implemented
// for performing an upgrade
type UpgradeTask interface {
	FromVersion() string
	ToVersion() string
	IsSuccess() error
	PreUpgrade() bool
	Upgrade() bool
	PostUpgrade() bool
}

// RunUpgrade runs all the upgrade tasks required
func RunUpgrade(tasks ...UpgradeTask) error {
	for _, task := range tasks {
		_ = task.PreUpgrade() && task.Upgrade() && task.PostUpgrade()
		if err := task.IsSuccess(); err != nil {
			return err
		}
	}
	return nil
}
