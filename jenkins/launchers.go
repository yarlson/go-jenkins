package jenkins

import (
	"encoding/xml"
	"fmt"
)

// Launcher is the interface for  all Jenkins node launchers.
type Launcher interface{}

// WorkDirSettings represents the Jenkins node work directory settings.
type WorkDirSettings struct {
	Disabled               bool   `json:"disabled" xml:"disabled"`
	InternalDir            string `json:"internalDir" xml:"internalDir"`
	FailIfWorkDirIsMissing bool   `json:"failIfWorkDirIsMissing" xml:"failIfWorkDirIsMissing"`
}

// JNLPLauncher represents a Jenkins JNLP launcher.
type JNLPLauncher struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`

	WebSocket       bool            `json:"websocket" xml:"websocket,omitempty"`
	WorkDirSettings WorkDirSettings `json:"workDirSettings,omitempty" xml:"workDirSettings,omitempty"`
}

// DefaultJNLPLauncher returns the default JNLP launcher.
func DefaultJNLPLauncher() *JNLPLauncher {
	return &JNLPLauncher{
		StaplerClass: "hudson.slaves.JNLPLauncher",
	}
}

// SSHHostKeyVerificationStrategy represents the Jenkins node SSH host key verification strategy.
type SSHHostKeyVerificationStrategy interface{}

// NonVerifyingKeyVerificationStrategy represents the Jenkins node non-verifying key verification strategy.
type NonVerifyingKeyVerificationStrategy struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`
}

func NewNonVerifyingKeyVerificationStrategy() *NonVerifyingKeyVerificationStrategy {
	return &NonVerifyingKeyVerificationStrategy{
		StaplerClass: "hudson.plugins.sshslaves.verifiers.NonVerifyingKeyVerificationStrategy",
	}
}

// KnownHostsFileKeyVerificationStrategy represents the Jenkins node known hosts file key verification strategy.
type KnownHostsFileKeyVerificationStrategy struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`
}

// ManuallyProvidedKeyVerificationStrategyKey represents the Jenkins node manually provided key verification strategy key.
type ManuallyProvidedKeyVerificationStrategyKey struct {
	Algorithm string `json:"algorithm" xml:"algorithm"`
	Key       string `json:"key" xml:"key"`
}

// ManuallyProvidedKeyVerificationStrategy represents the Jenkins node manually provided key verification strategy.
type ManuallyProvidedKeyVerificationStrategy struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`

	Key ManuallyProvidedKeyVerificationStrategyKey `json:"key" xml:"key"`
}

// ManuallyTrustedKeyVerificationStrategy represents the Jenkins node manually trusted key verification strategy.
type ManuallyTrustedKeyVerificationStrategy struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`

	RequireInitialManualTrust bool `json:"requireInitialManualTrust,omitempty" xml:"requireInitialManualTrust,omitempty"`
}

// SSHLauncher represents a Jenkins SSH launcher.
type SSHLauncher struct {
	StaplerClass string `json:"stapler-class" xml:"class,attr"`

	Host                 string `json:"host" xml:"host"`
	Port                 int    `json:"port" xml:"port"`
	CredentialID         string `json:"credentialId" xml:"credentialId"`
	LaunchTimeoutSeconds int    `json:"launchTimeoutSeconds" xml:"launchTimeoutSeconds"`
	MaxNumRetries        int    `json:"maxNumRetries" xml:"maxNumRetries"`
	RetryWaitTime        int    `json:"retryWaitTime" xml:"retryWaitTime"`
	TCPNoDelay           bool   `json:"tcpNoDelay" xml:"tcpNoDelay"`

	SSHHostKeyVerificationStrategy SSHHostKeyVerificationStrategy `json:"sshHostKeyVerificationStrategy" xml:"sshHostKeyVerificationStrategy"`
}

func NewSSHLauncher(host string, port int, credentialID string, launchTimeoutSeconds int, maxNumRetries int, retryWaitTime int, TCPNoDelay bool, SSHHostKeyVerificationStrategy interface{}) *SSHLauncher {
	return &SSHLauncher{
		StaplerClass:                   "hudson.plugins.sshslaves.SSHLauncher",
		Host:                           host,
		Port:                           port,
		CredentialID:                   credentialID,
		LaunchTimeoutSeconds:           launchTimeoutSeconds,
		MaxNumRetries:                  maxNumRetries,
		RetryWaitTime:                  retryWaitTime,
		TCPNoDelay:                     TCPNoDelay,
		SSHHostKeyVerificationStrategy: SSHHostKeyVerificationStrategy,
	}
}

// UnmarshalXML implements the xml.Unmarshaler interface.
// It decodes the XML attributes into the corresponding struct fields.
// It also decodes the XML child SSHHostKeyVerificationStrategy nodes into the corresponding struct fields.
func (n *SSHLauncher) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type Alias SSHLauncher // avoids recursive unmarshal
	v := &struct {
		SSHHostKeyVerificationStrategy struct {
			InnerXML []byte `xml:",innerxml"`  // Stores inner XML of the <SSHHostKeyVerificationStrategy> element
			Class    string `xml:"class,attr"` // Stores the class name from the <class> attribute
		} `xml:"sshHostKeyVerificationStrategy"`
		*Alias
	}{
		Alias: (*Alias)(n),
	}

	if err := d.DecodeElement(v, &start); err != nil {
		return err
	}

	// Converts InnerXML to a valid XMl document
	itemXML := []byte(fmt.Sprintf("<root>%s</root>", v.SSHHostKeyVerificationStrategy.InnerXML))

	switch v.SSHHostKeyVerificationStrategy.Class {
	case "hudson.plugins.sshslaves.verifiers.NonVerifyingKeyVerificationStrategy":
		n.SSHHostKeyVerificationStrategy = &NonVerifyingKeyVerificationStrategy{
			StaplerClass: v.SSHHostKeyVerificationStrategy.Class,
		}
	case "hudson.plugins.sshslaves.verifiers.KnownHostsFileKeyVerificationStrategy":
		n.SSHHostKeyVerificationStrategy = &KnownHostsFileKeyVerificationStrategy{
			StaplerClass: v.SSHHostKeyVerificationStrategy.Class,
		}
	case "hudson.plugins.sshslaves.verifiers.ManuallyProvidedKeyVerificationStrategy":
		n.SSHHostKeyVerificationStrategy = &ManuallyProvidedKeyVerificationStrategy{
			StaplerClass: v.SSHHostKeyVerificationStrategy.Class,
		}
		err := xml.Unmarshal(itemXML, n.SSHHostKeyVerificationStrategy)
		if err != nil {
			return err
		}
	case "hudson.plugins.sshslaves.verifiers.ManuallyTrustedKeyVerificationStrategy":
		n.SSHHostKeyVerificationStrategy = &ManuallyTrustedKeyVerificationStrategy{
			StaplerClass: v.SSHHostKeyVerificationStrategy.Class,
		}
		err := xml.Unmarshal(itemXML, n.SSHHostKeyVerificationStrategy)
		if err != nil {
			return err
		}
	}

	return nil
}
