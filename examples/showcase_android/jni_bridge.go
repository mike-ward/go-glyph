//go:build android

package main

/*
#include <jni.h>
#include <stdint.h>

// Go-exported functions (from main.go via //export).
// Types match the Go signatures on arm64: int→GoInt (long long),
// uintptr→GoUintptr (unsigned long long), float32→GoFloat32 (float).
extern void GlyphStart(unsigned long long windowPtr,
                        long long w, long long h, float scale);
extern void GlyphRender(long long w, long long h);
extern void GlyphScroll(float dy);
extern void GlyphTouch(float x, float y);
extern void GlyphResize(long long w, long long h);
extern void GlyphDestroy(void);

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_start(
    JNIEnv *env, jobject obj, jlong windowPtr,
    jint w, jint h, jfloat scale) {
    GlyphStart((unsigned long long)windowPtr,
               (long long)w, (long long)h, (float)scale);
}

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_render(
    JNIEnv *env, jobject obj, jint w, jint h) {
    GlyphRender((long long)w, (long long)h);
}

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_scroll(
    JNIEnv *env, jobject obj, jfloat dy) {
    GlyphScroll((float)dy);
}

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_touch(
    JNIEnv *env, jobject obj, jfloat x, jfloat y) {
    GlyphTouch((float)x, (float)y);
}

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_resize(
    JNIEnv *env, jobject obj, jint w, jint h) {
    GlyphResize((long long)w, (long long)h);
}

JNIEXPORT void JNICALL
Java_com_example_glyphshowcase_GlyphNative_destroy(
    JNIEnv *env, jobject obj) {
    GlyphDestroy();
}
*/
import "C"
