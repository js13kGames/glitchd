// Required at the very least on Windows, potentially other platforms, to be loaded before the
// import of grpc itself, due to the way its C bindings are constructed.
process.env['GRPC_SSL_CIPHER_SUITES'] = 'ECDHE-ECDSA-AES256-GCM-SHA384';

const
    ROOT_PATH          = __dirname + '/../../',
    VERSION            = require(ROOT_PATH + 'package').version,
    CONNECTION_TIMEOUT = 5, // Seconds.

    ADDR    = Symbol(),
    TOKEN   = Symbol(),
    OPTS    = Symbol(),
    SERVICE = Symbol(),

    grpc     = require('grpc'),
    services = grpc.load(ROOT_PATH + 'proto/items.proto').glitchd,

    defaultOpts = {
        serverCertFile:                    ROOT_PATH + 'server.crt',
        // clientKeyFile:                  undefined,
        // clientCertFile:                 undefined,
        'grpc.primary_user_agent':         'glitchd-client-node/' + VERSION,
        'grpc.max_send_message_length':    32 * 1024,   // Bytes.
        'grpc.max_receive_message_length': -1           // No limit.
    };

class ItemsStore {

    /**
     *
     * @param addr  string
     * @param token string
     * @param opts  Object
     */
    constructor (addr, token, opts) {
        if (typeof addr !== 'string') {
            throw new Error('Expected [addr] to be of type [string], got [' + typeof addr + '] instead.')
        }

        if (typeof token !== 'string') {
            throw new Error('Expected [token] to be of type [string], got [' + typeof token + '] instead.')
        }

        // Hardcoded server-side and invariant in length, so might as well avoid a potential roundtrip
        // for a failure.
        if (token.length !== 16) {
            throw new Error('Expected [token] to be 16 characters long, is [' + token.length + '] instead.')
        }

        if (opts !== undefined && !opts instanceof Object) {
            throw new Error('Expected [opts] to be an Object, got [' + typeof opts + '] instead.')
        }

        this[ADDR]  = addr;
        this[TOKEN] = token;
        this[OPTS]  = Object.assign({}, defaultOpts, opts || {});
    }

    /**
     *
     * @param {grpc~Deadline}   deadline    When to stop waiting for a connection to glitchd.
     * @return Promise Resolves to this ItemsStoreClient.
     */
    connect (deadline) {
        if (this.isReady()) {
            throw new Error('Already connected and ready.')
        }

        if (deadline === undefined) {
            deadline = new Date();
            deadline.setSeconds(deadline.getSeconds() + CONNECTION_TIMEOUT);
        } else if (!deadline instanceof Date && !deadline instanceof Number) {
            throw new Error('Expected [deadline] to be an instance of Number|Date, got ' + typeof deadline + ' instead.')
        }

        this[SERVICE] = new services.items.Store(this[ADDR], this.createCredentials(), this[OPTS]);

        return new Promise((resolve, reject) => {
            this[SERVICE].waitForReady(deadline, err => {
                if (err) {
                    reject(err)
                }
                resolve(this)
            });
        });
    }

    /**
     *
     * @param   key     string
     * @return  Promise
     */
    get (key) {
        return this.call('get', {key})
    }

    /**
     *
     * @param   key     string
     * @param   value   Buffer
     * @return  Promise
     */
    put (key, value) {
        if (!value instanceof Buffer) {
            throw new Error("Values passed to put must be instances of Buffer.")
        }

        return this.call('put', {key, value})
    }

    /**
     *
     * @param   key     string
     * @return  Promise
     */
    delete (key) {
        return this.call('delete', {key})
    }

    /**
     *
     * @return {grpc~Credentials}
     */
    createCredentials () {
        // For TLS, at least in case of glitchd, we need the bundled server's public key.
        // Will require a change when the key is to be fetched from a CA.
        if (typeof this[OPTS].serverCertFile !== 'string') {
            return grpc.credentials.createInsecure()
        }

        const fs = require('fs');

        return grpc.credentials.createSsl(
            fs.readFileSync(this[OPTS].serverCertFile),
            this[OPTS].clientKeyFile  !== undefined ? fs.readFileSync(this[OPTS].clientKeyFile)  : undefined,
            this[OPTS].clientCertFile !== undefined ? fs.readFileSync(this[OPTS].clientCertFile) : undefined
        );
    }

    /**
     *
     * @return {grpc~Metadata}
     */
    createFreshMetadata () {
        let md = new grpc.Metadata();
        md.add('token', this[TOKEN]);

        return md
    }

    /**
     *
     * @return  bool
     */
    isReady () {
        return this[SERVICE] !== undefined && this[SERVICE].$channel.getConnectivityState(true) === grpc.connectivityState.READY
    }

    /**
     *
     * @param   method  string
     * @param   message Object
     * @return  Promise
     */
    call (method, message) {
        return new Promise((resolve, reject) => {
            this[SERVICE][method](message, this.createFreshMetadata(), (err, response) => {
                if (err) {
                    reject(err)
                }
                resolve(response)
            })
        })
    }
}

// Note: Exporting as an object as other services will likely get added in the future.
module.exports = {
    ItemsStore
};
