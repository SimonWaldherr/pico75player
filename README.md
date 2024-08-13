# pico75player

play animated GIFs on a Hub75 Matrix with Raspberry Pi Pico (Pimoroni Interstate 75)  

convert GIF to SAG
```sh
go run gif2sag.go imgcolor/example.gif output.sag gif
```

convert it back to GIF (to check)
```sh
go run sag2gif.go output.sag output.gif
```

upload *output.sag* and *play_sag_on_hub75.py* to Raspberry Pi Pico and run the code

