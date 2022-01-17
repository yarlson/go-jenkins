package jenkins

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

func (s *Suite) TestNodeFillInNodeDefaults() {
	n := &Node{}
	n.fillInNodeDefaults()

	s.Equal(DefaultJNLPLauncher(), n.Launcher)
	s.Equal(DefaultNodeProperties(), n.Properties)
	s.Equal(DefaultNodeType(), n.Type)
	s.Equal(1, n.NumExecutors)
	s.Equal(DefaultRetentionsStrategy(), n.RetentionsStrategy)
}

func (s *Suite) TestNodesServiceCreate() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesCreateURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
		_, err := w.Write([]byte(
			`{"A":"B"}`,
		))
		s.NoError(err)
	})

	_, _, err = client.Nodes.Create(context.Background(), &Node{})
	s.NoError(err)
}

func (s *Suite) TestNodesServiceCreateError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	_, _, err = client.Nodes.Create(context.Background(), &Node{})
	s.Error(err)
}

func (s *Suite) TestNodesServiceList() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesListURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"computer":[{"displayName": "test"}]}`,
		))
		s.NoError(err)
	})
	nodes, _, err := client.Nodes.List(context.Background())
	s.NoError(err)
	s.Equal([]Node{{Name: "test"}}, nodes)
}

func (s *Suite) TestNodesServiceListError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	//lint:ignore SA1012 this is a test
	_, _, err = client.Nodes.List(nil)
	s.Error(err)
}

func (s *Suite) TestNodesServiceListUnmarshalError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(NodesListURL, func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`{"computer":[{"displayName": `,
		))
		s.NoError(err)
	})
	_, _, err = client.Nodes.List(context.Background())
	s.Error(err)
}

func (s *Suite) TestNodesServiceGet() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`
<?xml version="1.1" encoding="UTF-8"?>
<slave>
  <name>test</name>
  <description></description>
  <remoteFS>/var/lib/jenkins</remoteFS>
  <numExecutors>1</numExecutors>
  <mode>EXCLUSIVE</mode>
  <retentionStrategy class="hudson.slaves.RetentionStrategy$Always"/>
  <launcher class="hudson.slaves.JNLPLauncher">
    <workDirSettings>
      <disabled>false</disabled>
      <internalDir>remoting</internalDir>
      <failIfWorkDirIsMissing>false</failIfWorkDirIsMissing>
    </workDirSettings>
    <webSocket>false</webSocket>
  </launcher>
  <label>test</label>
  <nodeProperties/>
</slave>
`,
		))
		s.NoError(err)
	})

	node, _, err := client.Nodes.Get(context.Background(), "test")

	s.NoError(err)
	s.Equal("test", node.Name)
	s.IsType(&JNLPLauncher{}, node.Launcher)
}

func (s *Suite) TestNodesServiceGetError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	_, _, err = client.Nodes.Get(context.Background(), "test")

	s.Error(err)
}

func (s *Suite) TestNodesServiceGetUnmarshalError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "GET")
		_, err := w.Write([]byte(
			`<?xml version="2.0" encoding="UTF-8"?>
<slave>

</slave>
`,
		))
		s.NoError(err)
	})

	_, _, err = client.Nodes.Get(context.Background(), "test")

	s.Error(err)
}

func (s *Suite) TestNodesServiceUpdate() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	s.addCrumbsHandle()

	s.mux.HandleFunc(fmt.Sprintf(NodesGetURL, "test"), func(w http.ResponseWriter, r *http.Request) {
		s.testMethod(r, "POST")
	})

	_, _, err = client.Nodes.Update(context.Background(), &Node{
		Name:        "test",
		Description: "",
		RemoteFS:    "/var/lib/jenkins",
		Mode:        NodeModeExclusive,
		Labels:      []string{"test"},
	})

	s.NoError(err)
}

func (s *Suite) TestNodesServiceUpdateError() {
	s.newMux()
	client, err := NewClient(WithBaseURL(s.server.URL), WithPassword("admin", "admin"))
	s.NoError(err)

	_, _, err = client.Nodes.Update(context.Background(), &Node{
		Name:        "test",
		Description: "",
		RemoteFS:    "/var/lib/jenkins",
		Mode:        NodeModeExclusive,
		Labels:      []string{"test"},
	})

	s.Error(err)
}

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
