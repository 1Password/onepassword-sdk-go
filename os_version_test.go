package onepassword

import "testing"

func TestWindowsVersionFromVerOutput(t *testing.T) {
	input := "Microsoft Windows XP [Version 5.1.2600]"
	expectedOutput := "5.1.2600"

	output := windowsVersionFromVerOutput(input)

	if output != expectedOutput {
		t.Errorf("windowsVersionFromVerOutput(%q) returned %q and expected %q", input, output, expectedOutput)
	}
}

func TestLinuxVersionFromOSReleaseOutput(t *testing.T) {
	input := `NAME="Ubuntu"
VERSION="17.10 (Artful Aardvark)"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="Ubuntu 17.10"
VERSION_ID="17.10"
HOME_URL="https://www.ubuntu.com/"
SUPPORT_URL="https://help.ubuntu.com/"
BUG_REPORT_URL="https://bugs.launchpad.net/ubuntu/"
PRIVACY_POLICY_URL="https://www.ubuntu.com/legal/terms-and-policies/privacy-policy"
VERSION_CODENAME=artful
UBUNTU_CODENAME=artful
	`
	expectedOutput := "17.10"

	output := linuxVersionFromOSReleaseOutput(input)

	if output != expectedOutput {
		t.Errorf("linuxVersionFromOSReleaseOutput(%q) returned %q and expected %q", input, output, expectedOutput)
	}
}
