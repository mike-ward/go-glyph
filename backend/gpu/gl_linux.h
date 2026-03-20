//go:build linux && !android

#ifndef GL_LINUX_H
#define GL_LINUX_H

#include <stdint.h>
#include <stdlib.h>
#include "SDL.h"

// Opaque OpenGL context — defined in gl_linux.c.
typedef struct GLCtx GLCtx;

// Packed draw command matching Go drawCmd layout.
typedef struct {
	uint64_t textureID;
	int32_t  firstVert;
	int32_t  vertCount;
} CDrawCmd;

GLCtx*    glCtxInit(void *sdlWindow, float dpiScale);
uint64_t  glCtxNewTex(GLCtx *ctx, int w, int h);
void      glCtxUpdateTex(GLCtx *ctx, uint64_t tid,
                         void *data, int w, int h);
void      glCtxDeleteTex(GLCtx *ctx, uint64_t tid);
int       glCtxRender(GLCtx *ctx,
                      void *verts, int vertCount,
                      void *cmds,  int cmdCount,
                      float clearR, float clearG,
                      float clearB, float clearA,
                      int logicalW, int logicalH);
void      glCtxDestroy(GLCtx *ctx);
void      glCtxGetDrawableSize(GLCtx *ctx, int *w, int *h);

#endif
