# Fyne Apps Installer

This app is designed to help you browse and install available fyne apps.
It loads data using the [apps.fyne.io](https://apps.fyne.io) API to stay up to date.

This app builds all others from source code, so it can take some time.
It will also require some setup, shown in *Installing*, below.

## Installing

The easiest way to install this app is using the `fyne` command and it's `get` utility.
To do so you will need to have Go and C compilers installed - if you need help you can see the [fyne prerequisites](https://developer.fyne.io/started/).
Once set up execute the following

```
$ go get fyne.io/fyne/cmd/fyne
$ fyne get fyne.io/apps
```

The commands above will install the `Apps` app in your standard system location.
You can then launch it like any other app.

## Running

Once installed launch the app from your app launcher, file browser or programme manager by tapping on the icon:

<p style="align: center">
<img src="/Icon.png" width="256" />
</p>

When loaded you should see something like the following screenshot:

![](/img/screenshot.png)

Browse the apps, find one you like and tap the *Install* button.
Installed apps will appear alongside this app in the standard system location.
