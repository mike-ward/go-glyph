import UIKit
import Metal
import QuartzCore

class GlyphViewController: UIViewController {
    private var metalLayer: CAMetalLayer!
    private var displayLink: CADisplayLink?
    private var started = false

    override func loadView() {
        let v = UIView()
        v.backgroundColor = .black
        self.view = v
    }

    override func viewDidLoad() {
        super.viewDidLoad()

        guard let device = MTLCreateSystemDefaultDevice() else {
            fatalError("Metal not available")
        }

        metalLayer = CAMetalLayer()
        metalLayer.device = device
        metalLayer.pixelFormat = .bgra8Unorm
        metalLayer.contentsScale = UIScreen.main.scale
        metalLayer.framebufferOnly = true
        view.layer.addSublayer(metalLayer)

        let pan = UIPanGestureRecognizer(
            target: self, action: #selector(handlePan(_:)))
        view.addGestureRecognizer(pan)
    }

    override func viewDidLayoutSubviews() {
        super.viewDidLayoutSubviews()
        let bounds = view.bounds
        metalLayer.frame = bounds
        metalLayer.drawableSize = CGSize(
            width: bounds.width * UIScreen.main.scale,
            height: bounds.height * UIScreen.main.scale)

        let w = GoInt(bounds.width)
        let h = GoInt(bounds.height)

        if !started {
            let layerPtr = Unmanaged.passUnretained(metalLayer)
                .toOpaque()
            GlyphStart(
                GoUintptr(Int(bitPattern: layerPtr)),
                w, h,
                Float(UIScreen.main.scale))
            started = true

            displayLink = CADisplayLink(
                target: self,
                selector: #selector(render))
            displayLink?.add(to: .main, forMode: .default)
        } else {
            GlyphResize(w, h)
        }
    }

    @objc private func render() {
        let bounds = view.bounds
        GlyphRender(
            GoInt(bounds.width),
            GoInt(bounds.height))
    }

    @objc private func handlePan(
        _ gesture: UIPanGestureRecognizer
    ) {
        let translation = gesture.translation(in: view)
        GlyphScroll(Float(-translation.y))
        gesture.setTranslation(.zero, in: view)
    }

    override func touchesBegan(
        _ touches: Set<UITouch>, with event: UIEvent?
    ) {
        if let touch = touches.first {
            let loc = touch.location(in: view)
            GlyphTouch(Float(loc.x), Float(loc.y))
        }
    }

    deinit {
        displayLink?.invalidate()
        GlyphDestroy()
    }
}
