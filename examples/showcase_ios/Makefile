SDK_PATH := $(shell xcrun --sdk iphonesimulator --show-sdk-path)
CC_PATH  := $(shell xcrun --sdk iphonesimulator --find clang)

.PHONY: build clean

build:
	CGO_ENABLED=1 GOOS=ios GOARCH=arm64 \
	  CC="$(CC_PATH)" \
	  CGO_CFLAGS="-isysroot $(SDK_PATH) -mios-simulator-version-min=17.0 -fembed-bitcode" \
	  CGO_LDFLAGS="-isysroot $(SDK_PATH) -mios-simulator-version-min=17.0" \
	  go build -buildmode=c-archive -tags ios -o ios/libglyph.a .
	xcodebuild -project ios/ShowcaseApp.xcodeproj \
	  -scheme ShowcaseApp \
	  -sdk iphonesimulator \
	  -destination 'platform=iOS Simulator,name=iPhone 17 Pro' \
	  build

clean:
	rm -f ios/libglyph.a ios/libglyph.h
	xcodebuild -project ios/ShowcaseApp.xcodeproj \
	  -scheme ShowcaseApp clean 2>/dev/null || true
