# Audio Control

* using Windows Core Audio COM interfaces

* `github.com/HomeDing/goding/internal/audio` provides an abstraction over the Windows audio interfaces
* `github.com/deploymenttheory/go-bindings-win32` provides the generated Win32 and COM bindings

## Windows OLE Interfaces

To use the window specific Core Audio API, the relevant OLE interfaces in use are shown
in the following diagram:

``` txt

device volume

Console.  --  Games, system notification sounds, and voice commands.
Multimedia. -- Music, movies, narration, and live music recording
Communication. -- Voice communications (talking to another person)

OLE
    -> IMMDeviceEnumerator
        [.EnumAudioEndpoints()]
        .GetDefaultAudioEndpoint()
        .GetDevice(id)
        -> IMMDevice (Audio Endpoint)
            -> IAudioEndpointVolume
                .GetMasterVolumeLevelScalar()
                .SetMasterVolumeLevelScalar()
            -> PropertyStore
            -> IAudioSessionManager2
                -> IAudioSessionEnumerator
                    .GetCount()

```

TODO:

dev(ice):name
app(lication):name
"" default device
app:top -- application with top window 

* set Default Output Device (string)
* set Default Input Device (string)
* set Mute on application (string)
* set Volume on Output Device (string)

## See Also

* [core audio APIs](https://learn.microsoft.com/en-us/windows/win32/coreaudio/core-audio-apis-in-windows-vista)
* [AudioManager.cs](https://gist.github.com/sverrirs/d099b34b7f72bb4fb386)
* [text](https://hackaday.com/2026/01/31/motorized-faders-make-an-awesome-volume-mixer-for-your-pc/)
* [text](https://www.midi-mixer.com/)
* [text](https://github.com/AndreMiras/pycaw)
* [Py Get Speakers, GetAllSessions](https://github.com/AndreMiras/pycaw/blob/develop/pycaw/utils.py)