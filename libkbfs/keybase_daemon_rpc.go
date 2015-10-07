package libkbfs

import (
	"github.com/keybase/client/go/client"
	"github.com/keybase/client/go/engine"
	"github.com/keybase/client/go/libkb"
	"github.com/keybase/client/go/logger"
	keybase1 "github.com/keybase/client/go/protocol"
	rpc "github.com/keybase/go-framed-msgpack-rpc"
	"golang.org/x/net/context"
)

// KeybaseDaemonRPC implements the KeybaseDaemon interface using RPC
// calls.
type KeybaseDaemonRPC struct {
	identify keybase1.IdentifyInterface
	session  keybase1.SessionInterface
	favorite keybase1.FavoriteInterface
	user     keybase1.UserInterface
	log      logger.Logger
}

var _ KeybaseDaemon = KeybaseDaemonRPC{}

// NewKeybaseDaemonRPC makes a new KeybaseDaemonRPC that makes RPC
// calls using the socket of the given context.
func NewKeybaseDaemonRPC(ctx *libkb.GlobalContext, log logger.Logger) (KeybaseDaemonRPC, error) {
	_, xp, err := ctx.GetSocket()
	if err != nil {
		return KeybaseDaemonRPC{}, err
	}

	srv := rpc.NewServer(xp, libkb.WrapError)

	protocols := []rpc.Protocol{
		client.NewLogUIProtocol(),
		client.NewIdentifyUIProtocol(),
	}

	for _, p := range protocols {
		if err := srv.Register(p); err != nil {
			if _, ok := err.(rpc.AlreadyRegisteredError); !ok {
				return KeybaseDaemonRPC{}, err
			}
		}
	}

	client := rpc.NewClient(xp, libkb.UnwrapError)
	identifyClient := keybase1.IdentifyClient{Cli: client}
	sessionClient := keybase1.SessionClient{Cli: client}
	favoriteClient := keybase1.FavoriteClient{Cli: client}
	userClient := keybase1.UserClient{Cli: client}
	return newKeybaseDaemonRPCWithInterfaces(
		identifyClient, sessionClient, favoriteClient, userClient, log), nil
}

// For testing.
func newKeybaseDaemonRPCWithInterfaces(
	identify keybase1.IdentifyInterface,
	session keybase1.SessionInterface,
	favorite keybase1.FavoriteInterface,
	user keybase1.UserInterface,
	log logger.Logger,
) KeybaseDaemonRPC {
	return KeybaseDaemonRPC{
		identify: identify,
		session:  session,
		favorite: favorite,
		user:     user,
		log:      log,
	}
}

func (k KeybaseDaemonRPC) filterKeys(ctx context.Context, uid keybase1.UID, keys []keybase1.PublicKey) ([]VerifyingKey, []CryptPublicKey, error) {
	var verifyingKeys []VerifyingKey
	var cryptPublicKeys []CryptPublicKey
	for _, publicKey := range keys {
		if len(publicKey.PGPFingerprint) > 0 {
			continue
		}
		// Import the KID to validate it.
		key, err := libkb.ImportKeypairFromKID(publicKey.KID)
		if err != nil {
			return nil, nil, err
		}
		if publicKey.IsSibkey {
			k.log.CDebugf(
				ctx, "got verifying key %s for user %s",
				key.VerboseDescription(), uid)
			verifyingKeys = append(
				verifyingKeys, VerifyingKey{key.GetKID()})
		} else {
			k.log.CDebugf(
				ctx, "got crypt public key %s for user %s",
				key.VerboseDescription(), uid)
			cryptPublicKeys = append(
				cryptPublicKeys, CryptPublicKey{key.GetKID()})
		}
	}
	return verifyingKeys, cryptPublicKeys, nil
}

// Identify implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) Identify(ctx context.Context, assertion string) (
	UserInfo, error) {
	arg := engine.IDEngineArg{UserAssertion: assertion}.Export()
	var res keybase1.IdentifyRes
	f := func() error {
		var err error
		res, err = k.identify.Identify(arg)
		return err
	}
	if err := runUnlessCanceled(ctx, f); err != nil {
		return UserInfo{}, err
	}

	name := libkb.NewNormalizedUsername(res.User.Username)
	uid := keybase1.UID(res.User.Uid)

	verifyingKeys, cryptPublicKeys, err := k.filterKeys(ctx, uid, res.PublicKeys)
	if err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		Name:            name,
		UID:             uid,
		VerifyingKeys:   verifyingKeys,
		CryptPublicKeys: cryptPublicKeys,
	}, nil
}

// LoadUserPlusKeys implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) LoadUserPlusKeys(ctx context.Context, uid keybase1.UID) (
	UserInfo, error) {
	arg := keybase1.LoadUserPlusKeysArg{Uid: uid, CacheOK: true}
	var res keybase1.UserPlusKeys
	f := func() error {
		var err error
		res, err = k.user.LoadUserPlusKeys(arg)
		return err
	}
	if err := runUnlessCanceled(ctx, f); err != nil {
		return UserInfo{}, err
	}

	verifyingKeys, cryptPublicKeys, err := k.filterKeys(ctx, uid, res.DeviceKeys)
	if err != nil {
		return UserInfo{}, err
	}

	return UserInfo{
		Name:            libkb.NormalizedUsername(res.Username),
		UID:             res.Uid,
		VerifyingKeys:   verifyingKeys,
		CryptPublicKeys: cryptPublicKeys,
	}, nil
}

// CurrentUID implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) CurrentUID(ctx context.Context, sessionID int) (
	keybase1.UID, error) {
	var currentUID keybase1.UID
	f := func() error {
		var err error
		currentUID, err = k.session.CurrentUID(sessionID)
		return err
	}
	if err := runUnlessCanceled(ctx, f); err != nil {
		return keybase1.UID(""), err
	}
	return currentUID, nil
}

// CurrentSession implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) CurrentSession(ctx context.Context, sessionID int) (
	SessionInfo, error) {
	var res keybase1.Session
	f := func() error {
		var err error
		res, err = k.session.CurrentSession(sessionID)
		return err
	}
	if err := runUnlessCanceled(ctx, f); err != nil {
		return SessionInfo{}, err
	}
	// Import the KID to validate it.
	deviceSubkey, err := libkb.ImportKeypairFromKID(res.DeviceSubkeyKid)
	if err != nil {
		return SessionInfo{}, err
	}
	k.log.CDebugf(ctx, "got device subkey %s",
		deviceSubkey.GetKID().ToShortIDString())
	cryptPublicKey := CryptPublicKey{deviceSubkey.GetKID()}
	return SessionInfo{
		UID:            keybase1.UID(res.Uid),
		Token:          res.Token,
		CryptPublicKey: cryptPublicKey,
	}, nil
}

// FavoriteAdd implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) FavoriteAdd(ctx context.Context, folder keybase1.Folder) error {
	f := func() error {
		return k.favorite.FavoriteAdd(keybase1.FavoriteAddArg{Folder: folder})
	}
	return runUnlessCanceled(ctx, f)
}

// FavoriteDelete implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) FavoriteDelete(ctx context.Context, folder keybase1.Folder) error {
	f := func() error {
		return k.favorite.FavoriteDelete(keybase1.FavoriteDeleteArg{Folder: folder})
	}
	return runUnlessCanceled(ctx, f)
}

// FavoriteList implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) FavoriteList(ctx context.Context, sessionID int) ([]keybase1.Folder, error) {
	var folders []keybase1.Folder
	f := func() error {
		var err error
		folders, err = k.favorite.FavoriteList(sessionID)
		return err
	}
	if err := runUnlessCanceled(ctx, f); err != nil {
		return nil, err
	}
	return folders, nil
}

// Shutdown implements the KeybaseDaemon interface for KeybaseDaemonRPC.
func (k KeybaseDaemonRPC) Shutdown() {
	// Nothing to do.
}
