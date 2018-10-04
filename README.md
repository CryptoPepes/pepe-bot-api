# API for cryptopepe telegram bot

This is a simple but powerful api, which serves PNG files on request.
Results are based on random pepe DNA, unless a specific list of one or more properties is supplied,
 this will overwrite the corresponding properties in the randomly generated look.
 
Usage note: after overwriting properties, conflicts between parts are automatically resolved,
 which may result in a loss of some of the properties.
This can be resolved by supplying "None" to the conflicting properties, to explicitly disable them.


## Getting started

[`h2non/bimg`](https://github.com/h2non/bimg) is used for SVG -> PNG conversion.
This conversion requires `libvips` (Just `vips` on some systems) to be installed.

Running:

`go run *.go`


## Usage

API Route:
```
/profilepic?param1=...&param2=.....&etc....
```

Look params:
```
eye-color
eye-type
mouth
hair-color
hat-color
hat-color2
hair-type
neck
shirt-color
shirt-type
glasses-primary-color
glasses-secondary-colo
glasses-type
skin-color
background-color

```


### Color

Everything suffixed with "color" is a color, 
formatted as a 6 character RGB hex code, without the `#`. E.g. `660055`: 


### Size

The size of the image can be changed with the `size` param
 (it's a square image, hence only one param).

 
### Properties

Either the ID of a property, the name,
 or a fuzzy name (typo, partial, etc.) can be used to identify properties.

The names & IDs can be found in `mappings.json`


### RNG seed

Optionally, you can set the rng seed to use. 
This will determine the randomness: the same seed always results in the same image.

Param: `seed`, value: 64 bit number (dec base).


### Formats

The api defaults to PNG. The format can be changed to SVG using `format=svg`.
No other formats than SVG and PNG are supported at this time.


### Example

Example request:

```
/profilepic?seed=123&hair-type=bitcoin&glasses-type=sun1&neck=dollar&hat-color=660055&size=1000
```


## Vendoring:

Cryptopepe packages to add from local GO PATH:

```
govendor add cryptopepe.io/cryptopepe-reader/pepe
govendor add cryptopepe.io/cryptopepe-svg/builder/look
```

Other packages can be found automatically, simply run `govendor add +e` to add external packages used in the codebase.
Add the `-n` flag to dry-run.

When vendoring, don't forget to explicitly add the resource folder for svg template files:

```
govendor add cryptopepe.io/cryptopepe-svg/builder/tmpl/^
``` 


## Setting up a dokku app to deploy too

On remote:

```
# Make swap 
cd /var
touch swap.img
chmod 600 swap.img
dd if=/dev/zero of=/var/swap.img bs=1024k count=4000
mkswap /var/swap.img
swapon /var/swap.img
free
echo "/var/swap.img    none    swap    sw    0    0" >> /etc/fstab

# Create app
dokku apps:create botapi

# Set config options for SVG templating code
dokku config:set botapi APP_PATH="/app"

# Set the domain for the dokku app
dokku domains:add botapi botapi.cryptopepes.io

# Add letsencrypt plugin, for self-auto-renewing cert, HTTPS!
sudo dokku plugin:install https://github.com/dokku/dokku-letsencrypt.git

# Set an email for letsencrypt
dokku config:set --no-restart botapi DOKKU_LETSENCRYPT_EMAIL=cryptopepes@gmail.com

# Now enable letsencrypt!
dokku letsencrypt botapi
```

Now add the dokku instance as remote target and deploy:
On local:
```
git remote add ocean-botapi dokku@<IP-ADDRESS>:botapi

git push ocean-botapi master
```

On droplet after its running:
```
# remove default config (80:5000), it will auto-find the correct one (80:80)
dokku proxy:ports-remove botapi 80

# check result
dokku proxy:ports botapi
```
