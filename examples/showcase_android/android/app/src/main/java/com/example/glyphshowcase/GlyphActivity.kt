package com.example.glyphshowcase

import android.app.Activity
import android.opengl.GLSurfaceView
import android.os.Bundle
import android.view.MotionEvent
import javax.microedition.khronos.egl.EGLConfig
import javax.microedition.khronos.opengles.GL10

class GlyphActivity : Activity() {
    private lateinit var glView: GLSurfaceView
    @Volatile private var initialized = false
    private var lastTouchY = 0f
    private var surfaceWidth = 0
    private var surfaceHeight = 0

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        glView = GLSurfaceView(this).apply {
            setEGLContextClientVersion(3)
            setRenderer(object : GLSurfaceView.Renderer {
                override fun onSurfaceCreated(
                    gl: GL10?, config: EGLConfig?
                ) {
                    val dm = resources.displayMetrics
                    val scale = dm.density
                    val w = dm.widthPixels
                    val h = dm.heightPixels
                    // Note: nativeWindow is obtained from the
                    // Surface by the Go c-shared library via EGL
                    // init. GlyphStart is called with the
                    // ANativeWindow pointer passed from JNI.
                    GlyphNative.start(0, w, h, scale)
                    initialized = true
                }

                override fun onSurfaceChanged(
                    gl: GL10?, width: Int, height: Int
                ) {
                    surfaceWidth = width
                    surfaceHeight = height
                    if (initialized) {
                        GlyphNative.resize(width, height)
                    }
                }

                override fun onDrawFrame(gl: GL10?) {
                    if (initialized) {
                        GlyphNative.render(
                            surfaceWidth, surfaceHeight
                        )
                    }
                }
            })
            renderMode = GLSurfaceView.RENDERMODE_WHEN_DIRTY
        }

        setContentView(glView)
    }

    override fun onTouchEvent(event: MotionEvent): Boolean {
        when (event.action) {
            MotionEvent.ACTION_DOWN -> {
                lastTouchY = event.y
                GlyphNative.touch(event.x, event.y)
            }
            MotionEvent.ACTION_MOVE -> {
                val dy = lastTouchY - event.y
                lastTouchY = event.y
                if (initialized) GlyphNative.scroll(dy)
                GlyphNative.touch(event.x, event.y)
            }
        }
        glView.requestRender()
        return true
    }

    override fun onPause() {
        super.onPause()
        glView.onPause()
    }

    override fun onResume() {
        super.onResume()
        glView.onResume()
    }

    override fun onDestroy() {
        if (initialized) {
            glView.queueEvent { GlyphNative.destroy() }
        }
        super.onDestroy()
    }
}
