package sms // import "github.com/mndrix/sms-over-xmpp"
import xco "github.com/mndrix/go-xco"

// Config describes the minimum methods necessary for configuring an
// sms-over-xmpp component.  These are methods for which no sensible
// default is possible.  Optional configuration methods are described
// by other interfaces.
type Config interface {
	// AddressToPhone converts an XMPP address into an E.164 phone
	// number.  This determines the mapping from XMPP users to PSTN
	// users.
	//
	// Should return ErrIgnoreMessage if XMPP messages to this address
	// should be ignored completely.
	AddressToPhone(xco.Address) (string, error)

	// ComponentName is a name (usually a domain name) by which the
	// XMPP server knows us.
	ComponentName() string

	// HttpHost is the host address on which to listen for HTTP
	// connections.  If its the empty string, listen on all available
	// interfaces.
	HttpHost() string

	// HttpPort is the port on which to listen for HTTP connections.
	HttpPort() int

	// SharedSecret is the secret with which we can authenticate to
	// the XMPP server.
	SharedSecret() string

	// SmsProvider returns a provider that's able to send and receive
	// SMS messages to and from the numbers indicated.
	SmsProvider(from string, to string) (SmsProvider, error)

	// XmppHost is the domain name or IP address of the XMPP server.
	XmppHost() string

	// XmppPort is the port on which the XMPP server is listening.
	XmppPort() int
}