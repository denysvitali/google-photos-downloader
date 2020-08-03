# google-photos-downloader

A CLI to download your photos from Google Photos


```bash
Usage: photosdownloader [--client-id CLIENT-ID] [--client-secret CLIENT-SECRET] --output OUTPUT ALBUMID

Positional arguments:
  ALBUMID                Album ID

Options:
  --client-id CLIENT-ID
                         API's Client ID'
  --client-secret CLIENT-SECRET
                         API's Client Secret'
  --output OUTPUT, -o OUTPUT
  --help, -h             display this help and exit

```

E.g: 

```
$ export CLIENT_ID=XXXXX.apps.googleusercontent.com
$ export CLIENT_SECRET=Cp1d5VCTKYrKYdcyHy6H

$ photosdownloader -o ~/Photos "AFh5vF..."

$ ls ~/Photos | wc -l                                                                                                      -130-
987
```