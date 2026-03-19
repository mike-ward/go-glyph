package com.example.glyphshowcase

/**
 * JNI declarations for the Go c-shared library exports.
 */
object GlyphNative {
    init {
        System.loadLibrary("glyph")
    }

    external fun start(windowPtr: Long, w: Int, h: Int, scale: Float)
    external fun render(w: Int, h: Int)
    external fun scroll(dy: Float)
    external fun touch(x: Float, y: Float)
    external fun resize(w: Int, h: Int)
    external fun destroy()
}
