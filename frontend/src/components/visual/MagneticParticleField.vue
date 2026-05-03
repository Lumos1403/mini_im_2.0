<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

interface Particle {
  x: number
  y: number
  baseX: number
  baseY: number
  vx: number
  vy: number
  radius: number
  depth: number
  alpha: number
}

interface MagnetFocus {
  x: number
  y: number
  radius: number
  strength: number
}

const canvasRef = ref<HTMLCanvasElement | null>(null)

let ctx: CanvasRenderingContext2D | null = null
let particles: Particle[] = []
let animationFrame = 0
let width = 0
let height = 0
let pixelRatio = 1
let pointer: MagnetFocus | null = null
let buttonFocus: MagnetFocus | null = null
let lastPointerMove = 0

onMounted(() => {
  const canvas = canvasRef.value
  if (!canvas) {
    return
  }
  ctx = canvas.getContext('2d')
  resize()
  window.addEventListener('resize', resize)
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  window.addEventListener('pointerleave', clearPointer)
  window.addEventListener('auth-magnet-focus', handleButtonFocus as EventListener)
  window.addEventListener('auth-magnet-clear', clearButtonFocus)
  animationFrame = window.requestAnimationFrame(draw)
})

onBeforeUnmount(() => {
  window.cancelAnimationFrame(animationFrame)
  window.removeEventListener('resize', resize)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerleave', clearPointer)
  window.removeEventListener('auth-magnet-focus', handleButtonFocus as EventListener)
  window.removeEventListener('auth-magnet-clear', clearButtonFocus)
})

function resize() {
  const canvas = canvasRef.value
  if (!canvas) {
    return
  }

  const rect = canvas.getBoundingClientRect()
  width = Math.max(1, rect.width)
  height = Math.max(1, rect.height)
  pixelRatio = Math.min(window.devicePixelRatio || 1, 2)
  canvas.width = Math.floor(width * pixelRatio)
  canvas.height = Math.floor(height * pixelRatio)
  canvas.style.width = `${width}px`
  canvas.style.height = `${height}px`
  ctx?.setTransform(pixelRatio, 0, 0, pixelRatio, 0, 0)
  seedParticles()
}

function seedParticles() {
  const count = Math.min(190, Math.max(90, Math.floor((width * height) / 9500)))
  particles = Array.from({ length: count }, (_, index) => {
    const depth = 0.35 + Math.random() * 1.2
    const gridBias = index / count
    const x = (Math.random() * 0.9 + 0.05) * width
    const y = (Math.random() * 0.9 + 0.05) * height + Math.sin(gridBias * Math.PI * 4) * 12
    return {
      x,
      y,
      baseX: x,
      baseY: y,
      vx: 0,
      vy: 0,
      radius: 0.75 + depth * 1.6,
      depth,
      alpha: 0.22 + depth * 0.22,
    }
  })
}

function handlePointerMove(event: PointerEvent) {
  pointer = {
    x: event.clientX,
    y: event.clientY,
    radius: 280,
    strength: 0.95,
  }
  lastPointerMove = performance.now()
}

function clearPointer() {
  pointer = null
}

function handleButtonFocus(event: CustomEvent<MagnetFocus>) {
  buttonFocus = event.detail
}

function clearButtonFocus() {
  buttonFocus = null
}

function draw(now: number) {
  if (!ctx) {
    return
  }

  if (pointer && now - lastPointerMove > 1800) {
    pointer = null
  }

  ctx.clearRect(0, 0, width, height)
  paintAtmosphere()

  for (const particle of particles) {
    applyParticleMotion(particle, pointer)
    applyParticleMotion(particle, buttonFocus)

    const springX = (particle.baseX - particle.x) * 0.012 * particle.depth
    const springY = (particle.baseY - particle.y) * 0.012 * particle.depth
    particle.vx = (particle.vx + springX) * 0.9
    particle.vy = (particle.vy + springY) * 0.9
    particle.x += particle.vx
    particle.y += particle.vy

    const glow = particleGlow(particle)
    ctx.beginPath()
    ctx.fillStyle = `rgba(240, 207, 132, ${Math.min(0.9, particle.alpha + glow)})`
    ctx.shadowColor = 'rgba(240, 207, 132, 0.75)'
    ctx.shadowBlur = 10 * glow
    ctx.arc(particle.x, particle.y, particle.radius + glow * 1.8, 0, Math.PI * 2)
    ctx.fill()
  }

  ctx.shadowBlur = 0
  paintMagneticLines()
  animationFrame = window.requestAnimationFrame(draw)
}

function paintAtmosphere() {
  if (!ctx) {
    return
  }

  const base = ctx.createRadialGradient(width * 0.5, height * 0.2, 20, width * 0.5, height * 0.45, width * 0.9)
  base.addColorStop(0, 'rgba(118, 95, 52, 0.16)')
  base.addColorStop(0.5, 'rgba(24, 27, 28, 0.22)')
  base.addColorStop(1, 'rgba(0, 0, 0, 0)')
  ctx.fillStyle = base
  ctx.fillRect(0, 0, width, height)

  const focus = buttonFocus || pointer
  if (focus) {
    const halo = ctx.createRadialGradient(focus.x, focus.y, 0, focus.x, focus.y, focus.radius * 0.8)
    halo.addColorStop(0, 'rgba(240, 207, 132, 0.18)')
    halo.addColorStop(0.34, 'rgba(215, 180, 106, 0.07)')
    halo.addColorStop(1, 'rgba(215, 180, 106, 0)')
    ctx.fillStyle = halo
    ctx.fillRect(0, 0, width, height)
  }
}

function applyParticleMotion(particle: Particle, focus: MagnetFocus | null) {
  if (!focus) {
    return
  }
  const dx = focus.x - particle.x
  const dy = focus.y - particle.y
  const distSq = dx * dx + dy * dy
  const radiusSq = focus.radius * focus.radius
  if (distSq > radiusSq || distSq < 1) {
    return
  }
  const distance = Math.sqrt(distSq)
  const influence = (1 - distance / focus.radius) ** 2
  const pull = influence * focus.strength * particle.depth * 0.045
  const tangent = influence * 0.012
  particle.vx += dx * pull - dy * tangent
  particle.vy += dy * pull + dx * tangent
}

function particleGlow(particle: Particle) {
  const focus = buttonFocus || pointer
  if (!focus) {
    return 0
  }
  const dx = focus.x - particle.x
  const dy = focus.y - particle.y
  const distance = Math.sqrt(dx * dx + dy * dy)
  if (distance > focus.radius) {
    return 0
  }
  return (1 - distance / focus.radius) * 0.42
}

function paintMagneticLines() {
  if (!ctx || !pointer) {
    return
  }

  ctx.save()
  ctx.strokeStyle = 'rgba(240, 207, 132, 0.08)'
  ctx.lineWidth = 1
  for (let index = 0; index < 18; index += 1) {
    const angle = (Math.PI * 2 * index) / 18
    const radius = 54 + (index % 5) * 24
    ctx.beginPath()
    ctx.ellipse(pointer.x, pointer.y, radius * 1.55, radius * 0.34, angle, 0, Math.PI * 2)
    ctx.stroke()
  }
  ctx.restore()
}
</script>

<template>
  <canvas ref="canvasRef" class="magnetic-particle-field" aria-hidden="true"></canvas>
</template>

<style scoped>
.magnetic-particle-field {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

@media (prefers-reduced-motion: reduce) {
  .magnetic-particle-field {
    opacity: 0.35;
  }
}
</style>
