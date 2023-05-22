package flags

import (
	"errors"
	"flag"
)

var (
	NickFlag       *string
	DescFlag       *string
	TargetNickFlag *string
	WpbotFlag      *bool
)

func LoadFlags() {
	TargetNickFlag = flag.String("target_nick", "", "Specify the target nickname")
	NickFlag = flag.String("nick", "", "Specify your nickname")
	DescFlag = flag.String("desc", "", "Specify your description")
	WpbotFlag = flag.Bool("wpbot", false, "Specify if you want to play with WP bot")
	flag.Parse()
}

func ValidateFlags() error {
	if err := ValidateNick(); err != nil {
		return err
	}
	return nil
}

func ValidateNick() error {
	if len(*NickFlag) < 2 || len(*NickFlag) > 10 {
		return errors.New("TargetNick must be between 2 and 10 characters")
	}
	return nil
}
