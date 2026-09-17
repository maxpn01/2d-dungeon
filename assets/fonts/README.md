# Pixel Operator

Pixel Operator by Jayvee Enaguas (HarvettFox96), released under CC0 1.0.
The original license and downloaded font variants are in `pixel-operator/`.

The game embeds `PixelOperator.ttf` and `PixelOperator-Bold.ttf` through
`assets/fonts.go`, so no external asset directory is needed at runtime.
Gameplay text uses regular; text at 24 pixels and above uses bold.
Menu buttons (including game-over actions) use raylib's default font. The
main-menu and game-over titles retain their custom pixel lettering.

`internal/game/font.go` caches a font atlas for each requested size, uses point
filtering, and shares font metrics between drawing and layout. Use its
`drawText` and `measureText` helpers for new UI text.
