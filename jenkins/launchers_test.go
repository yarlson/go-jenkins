package jenkins

import "encoding/xml"

func (s *Suite) TestSSHLauncherMarshalNonVerifyingKeyVerificationStrategy() {
	inputXML := `<launcher class="hudson.plugins.sshslaves.SSHLauncher" plugin="ssh-slaves@1.33.0">
    <host>ss</host>
    <port>22</port>
    <credentialsId>ss</credentialsId>
    <launchTimeoutSeconds>60</launchTimeoutSeconds>
    <maxNumRetries>10</maxNumRetries>
    <retryWaitTime>15</retryWaitTime>
    <sshHostKeyVerificationStrategy class="hudson.plugins.sshslaves.verifiers.NonVerifyingKeyVerificationStrategy"/>
    <tcpNoDelay>true</tcpNoDelay>
  </launcher>`
	var launcher SSHLauncher
	err := xml.Unmarshal([]byte(inputXML), &launcher)
	s.NoError(err)
	s.Equal("hudson.plugins.sshslaves.SSHLauncher", launcher.StaplerClass)
	s.Equal("hudson.plugins.sshslaves.verifiers.NonVerifyingKeyVerificationStrategy", launcher.SSHHostKeyVerificationStrategy.(*NonVerifyingKeyVerificationStrategy).StaplerClass)
}

func (s *Suite) TestSSHLauncherMarshalKnownHostsFileKeyVerificationStrategy() {
	inputXML := `<launcher class="hudson.plugins.sshslaves.SSHLauncher" plugin="ssh-slaves@1.33.0">
    <host>ss</host>
    <port>22</port>
    <credentialsId>ss</credentialsId>
    <launchTimeoutSeconds>60</launchTimeoutSeconds>
    <maxNumRetries>10</maxNumRetries>
    <retryWaitTime>15</retryWaitTime>
    <sshHostKeyVerificationStrategy class="hudson.plugins.sshslaves.verifiers.KnownHostsFileKeyVerificationStrategy"/>
    <tcpNoDelay>true</tcpNoDelay>
  </launcher>`
	var launcher SSHLauncher
	err := xml.Unmarshal([]byte(inputXML), &launcher)
	s.NoError(err)
	s.Equal("hudson.plugins.sshslaves.SSHLauncher", launcher.StaplerClass)
	s.Equal("hudson.plugins.sshslaves.verifiers.KnownHostsFileKeyVerificationStrategy", launcher.SSHHostKeyVerificationStrategy.(*KnownHostsFileKeyVerificationStrategy).StaplerClass)
}

func (s *Suite) TestSSHLauncherMarshalManuallyProvidedKeyVerificationStrategy() {
	inputXML := `<launcher class="hudson.plugins.sshslaves.SSHLauncher" plugin="ssh-slaves@1.33.0">
    <host>ss</host>
    <port>22</port>
    <credentialsId>ss</credentialsId>
    <launchTimeoutSeconds>60</launchTimeoutSeconds>
    <maxNumRetries>10</maxNumRetries>
    <retryWaitTime>15</retryWaitTime>
    <sshHostKeyVerificationStrategy class="hudson.plugins.sshslaves.verifiers.ManuallyProvidedKeyVerificationStrategy">
      <key>
        <algorithm>ssh-rsa</algorithm>
        <key>AAAAB3NzaC1yc2EAAAADAQABAAABAQDoNycc11khfOqTtpnOFq3MR9r24R/4s6lAoCbBLIMJ+1GlB4qaWLJg6Me1RCuBovvZMvpxJvDZHw8cgFrPFFHw029VtCBVH0e1ifSWpQREYk2GpL0jdfFzkavxHmWTlu1HXvK5Q9vwqCAuq1ZSKza28J26ZY7vhwgjY+25o18gswR2omLkYVDBo0N2REZ6pQqpUTNfsfFgJ0mGsgRYOPdtx0TiMskCggz8xl/11QIohEwauT2nt8+fpJGAU8JO4JrWB7LNzIBLEL+Uk2ZgK/VEbUIH6Dn9mCwEiztWQ3XnXJ0TcZ/MVeaQUby+MKMShk1JHrsTqJygQLDb7SQ2X+4j</key>
      </key>
    </sshHostKeyVerificationStrategy>
    <tcpNoDelay>true</tcpNoDelay>
  </launcher>`
	var launcher SSHLauncher
	err := xml.Unmarshal([]byte(inputXML), &launcher)
	s.NoError(err)
	s.Equal("hudson.plugins.sshslaves.SSHLauncher", launcher.StaplerClass)
	s.Equal("hudson.plugins.sshslaves.verifiers.ManuallyProvidedKeyVerificationStrategy", launcher.SSHHostKeyVerificationStrategy.(*ManuallyProvidedKeyVerificationStrategy).StaplerClass)
}

func (s *Suite) TestSSHLauncherMarshalManuallyTrustedKeyVerificationStrategy() {
	inputXML := `<launcher class="hudson.plugins.sshslaves.SSHLauncher" plugin="ssh-slaves@1.33.0">
    <host>ss</host>
    <port>22</port>
    <credentialsId>ss</credentialsId>
    <launchTimeoutSeconds>60</launchTimeoutSeconds>
    <maxNumRetries>10</maxNumRetries>
    <retryWaitTime>15</retryWaitTime>
   <sshHostKeyVerificationStrategy class="hudson.plugins.sshslaves.verifiers.ManuallyTrustedKeyVerificationStrategy">
      <requireInitialManualTrust>true</requireInitialManualTrust>
    </sshHostKeyVerificationStrategy>
    <tcpNoDelay>true</tcpNoDelay>
    <tcpNoDelay>true</tcpNoDelay>
  </launcher>`
	var launcher SSHLauncher
	err := xml.Unmarshal([]byte(inputXML), &launcher)
	s.NoError(err)
	s.Equal("hudson.plugins.sshslaves.SSHLauncher", launcher.StaplerClass)
	s.Equal("hudson.plugins.sshslaves.verifiers.ManuallyTrustedKeyVerificationStrategy", launcher.SSHHostKeyVerificationStrategy.(*ManuallyTrustedKeyVerificationStrategy).StaplerClass)
	s.True(launcher.SSHHostKeyVerificationStrategy.(*ManuallyTrustedKeyVerificationStrategy).RequireInitialManualTrust)
}
