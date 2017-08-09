# glitchd-client-node

Node.js client for glitchd - a set of services made available to participants of the
[js13kgames](https://js13kgames.com) competition.

### Disclaimer
Glitchd is an **experimental** service. While we will try our best to
keep BC breaks at a minimum and the service as reliable as possible,
you should assume that it is not reliable and appropriately handle this.
Whether and/or how you handle this will reflect in your game's technical
score.

### Obtaining access
Glitchd uses multiplexed HTTP2 (over TLS) under the hood, but authorization
of clients is bearer token based to keep it a bit more familiar.

- Each participant may request one token per game/draft
  in the **server** category of the competition.
- You need to be registered on the js13kgames website.
- Tokens are handed out manually. Poke **@alkor** on the js13kgames Slack. The process will be automated
in the future, but is deliberately personal while the project is in development,
in case we need to keep participants up to speed about urgent changes
and/or resolve potential abuse.

Upon obtaining your token, expose it to the js13kgames Node sandbox
as a `GLITCHD_TOKEN` env variable. The sandbox will then take care of constructing
the glitchd client and exposing it to your game's code upon start.

On Heroku, please consult [this](https://devcenter.heroku.com/articles/config-vars)
tutorial on how to define env vars. On your local dev machine - it's up to you.

Do **not** expose your token publicly. Treat it as a secret credential.
In case of a suspected breach your token can be rotated.

### Limits
- We currently do not limit the number of requests per second made to
the service, but will be monitoring usage and adjusting this if needed.
As a good neighbour policy, try to stay below 100 RPS per token.
The code can handle magnitudes more, but our hardware resources are
very limited.
- The **Items** service has a max message size set to **32KiB**.
If you have a use case (game) that requires more - get in touch.
- We currently do not limit the total size of data in the **Items** service.
Consider 250MiB to be a soft limit, per token (eg. per game).
That's effectively ~8000 keys with max size values.
- Keep in mind that communication with glitchd will in most cases have
a severe network latency penalty, so even while the server will usually
complete your read request in <1ms and writes in <2ms, the roundtrip
to the server may take 200ms. Do not use it as a cache.
You've got Node.js - you've got memory for this.


## Available services
### Items
The items service is a persistent key-value store targetted at random reads
(eg. retrieving single items) of relatively small data (<32KiB per item).

The service is agnostic to what kind of data you put in, as long as it
gets to *om nom nom* its **bytes**. It does not impose any restrictions
on the format of keys, besides them being strings (which internally
are stored as byte arrays, just as the values). Feel free to be clever
and optimize your key patterns for lookups - especially when storing
many of them.

The service provides the following methods:

```javascript
get (key: string) : Promise
```
```javascript
put (key: string, value: Buffer) : Promise
```
```javascript
delete (key: string) : Promise
```


The client exposes some additional methods related to connectivity. Please
see the example below for a full flow, or consult the source.

##### Example usage (within the js13kgames Node sandbox)

*Note*: in the sandbox, the *Items* service is exposed as a global
`glitchd.items` object. The client gets constructed and exposed by the sandbox
**only** when a `GLITCHD_TOKEN` (see "Obtaining access" above)
environment variable is present and set.

```javascript
    // Connections to glitchd need to be made explicitly for the underlying client's
    // methods to be available. You can provide an optional connection deadline (Date)
    // as param to connect(). When omitted, the default of now + CONNECTION_TIMEOUT (5s) gets used.
    await glitchd.items.connect();

    // Set a value and then retrieve its contents from the remote store.
    await client.put('foo', Buffer.from('bar'));
    res = await glitchd.items.get('foo');

    // The value for the key will be a Buffer, just as you stored it,
    // and available as the "value" property on the response object.
    console.log(res.value.toString());
    // >> 'bar'

    // Delete a value from the store and then attempt to retrieve its contents.
    // Note: Deleting an item which is not set is a no-op (with no errors).
    await glitchd.items.delete('foo');

    // Trying to retrieve a value that does not exist does result in an error, however.
    try {
        await glitchd.items.get('foo');
    } catch(err) {
        console.log(err.code, err.message)
        // >> 5 'No value found for the requested key.'

        // Note: gRPC uses non-HTTP codes to simplify its protocol.
        // Code 5 corresponds to NotFound.
        // You can find the full enum of codes here:
        // https://github.com/grpc/grpc-go/blob/master/codes/codes.go
    }
```

##### Batch requests

Batching of reads or writes is currently not implemented but very much
possible over the protocol used. If you have a particular use case that
could benefit from this, please get in touch and we will do our best to
accommodate you.
