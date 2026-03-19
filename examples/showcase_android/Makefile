UNAME_S       := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  NDK_HOST    ?= darwin-x86_64
else
  NDK_HOST    ?= linux-x86_64
endif

NDK_TOOLCHAIN ?= $(ANDROID_NDK_HOME)/toolchains/llvm/prebuilt/$(NDK_HOST)
NDK_CLANG     ?= $(NDK_TOOLCHAIN)/bin/aarch64-linux-android24-clang
ANDROID_HOME  ?= $(HOME)/Library/Android/sdk
ADB           ?= $(ANDROID_HOME)/platform-tools/adb
GRADLEW       := android/gradlew
GRADLE_JAR    := android/gradle/wrapper/gradle-wrapper.jar

.PHONY: build clean check-ndk apk install run setup

check-ndk:
ifndef ANDROID_NDK_HOME
	$(error ANDROID_NDK_HOME is not set. Install the Android NDK and export ANDROID_NDK_HOME=/path/to/ndk)
endif
	@test -x "$(NDK_CLANG)" || \
	  (echo "error: NDK clang not found at $(NDK_CLANG)" && exit 1)

build: check-ndk
	CGO_ENABLED=1 GOOS=android GOARCH=arm64 \
	  CC="$(NDK_CLANG)" \
	  go build -buildmode=c-shared -tags android \
	  -o android/app/src/main/jniLibs/arm64-v8a/libglyph.so .

$(GRADLE_JAR):
	@echo "Generating Gradle wrapper (requires 'gradle' on PATH)..."
	@command -v gradle >/dev/null 2>&1 || \
	  (echo "error: gradle not found. Install with: brew install gradle" && exit 1)
	cd android && gradle wrapper --gradle-version 8.14

setup: $(GRADLE_JAR)
	@echo "Gradle wrapper ready."

apk: build $(GRADLE_JAR)
	cd android && ./gradlew assembleDebug

install: build $(GRADLE_JAR)
	cd android && ./gradlew installDebug

run: install
	$(ADB) shell am start -n com.example.glyphshowcase/.GlyphActivity

clean:
	rm -f android/app/src/main/jniLibs/arm64-v8a/libglyph.so
	cd android && [ -f gradlew ] && ./gradlew clean || true
