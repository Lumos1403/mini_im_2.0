<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

interface Particle {
  x: number
  y: number
  vx: number
  vy: number
  radius: number
  baseAlpha: number
  depth: number
}

const x = ref(50)
const y = ref(35)
const canvasRef = ref<HTMLCanvasElement | null>(null)

let ctx: CanvasRenderingContext2D | null = null
let particles: Particle[] = []
let animationFrame = 0
let width = 0
let height = 0
let pixelRatio = 1
let pointerX = 0
let pointerY = 0
let pointerActive = false
let lastPointerMove = 0

const pointerIdleTimeout = 1500
const influenceRadius = 220

function handlePointerMove(event: PointerEvent) {
  x.value = (event.clientX / Math.max(window.innerWidth, 1)) * 100
  y.value = (event.clientY / Math.max(window.innerHeight, 1)) * 100
  pointerX = event.clientX
  pointerY = event.clientY
  pointerActive = true
  lastPointerMove = performance.now()
}

function clearPointer() {
  pointerActive = false
}

function resizeCanvas() {
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
  const count = Math.min(150, Math.max(64, Math.floor((width * height) / 14000)))
  particles = Array.from({ length: count }, () => {
    const depth = 0.45 + Math.random() * 0.9
    return {
      x: Math.random() * width,
      y: Math.random() * height,
      vx: (Math.random() - 0.5) * 0.16,
      vy: (Math.random() - 0.5) * 0.16,
      radius: 0.85 + depth * 1.35,
      baseAlpha: 0.1 + depth * 0.09,
      depth,
    }
  })
}

function draw(now: number) {
  if (!ctx) {
    return
  }

  if (pointerActive && now - lastPointerMove > pointerIdleTimeout) {
    pointerActive = false
  }

  ctx.clearRect(0, 0, width, height)

  for (const particle of particles) {
    moveParticle(particle)
    paintParticle(particle)
  }

  animationFrame = window.requestAnimationFrame(draw)
}

function moveParticle(particle: Particle) {
  particle.vx += (Math.random() - 0.5) * 0.024 * particle.depth
  particle.vy += (Math.random() - 0.5) * 0.024 * particle.depth
  particle.vx *= 0.988
  particle.vy *= 0.988
  particle.x += particle.vx
  particle.y += particle.vy

  if (particle.x < -8) {
    particle.x = width + 8
  } else if (particle.x > width + 8) {
    particle.x = -8
  }

  if (particle.y < -8) {
    particle.y = height + 8
  } else if (particle.y > height + 8) {
    particle.y = -8
  }
}

function paintParticle(particle: Particle) {
  if (!ctx) {
    return
  }

  const proximity = particleProximity(particle)
  const alpha = particle.baseAlpha + proximity * 0.44
  const radius = particle.radius + proximity * 1.7

  if (proximity > 0.28) {
    const haloRadius = radius * (9 + proximity * 9)
    const halo = ctx.createRadialGradient(particle.x, particle.y, 0, particle.x, particle.y, haloRadius)
    halo.addColorStop(0, `rgba(240, 207, 132, ${0.22 * proximity})`)
    halo.addColorStop(0.42, `rgba(215, 180, 106, ${0.08 * proximity})`)
    halo.addColorStop(1, 'rgba(215, 180, 106, 0)')
    ctx.fillStyle = halo
    ctx.beginPath()
    ctx.arc(particle.x, particle.y, haloRadius, 0, Math.PI * 2)
    ctx.fill()
  }

  ctx.fillStyle = `rgba(240, 207, 132, ${Math.min(alpha, 0.68)})`
  ctx.beginPath()
  ctx.arc(particle.x, particle.y, radius, 0, Math.PI * 2)
  ctx.fill()
}

function particleProximity(particle: Particle) {
  if (!pointerActive) {
    return 0
  }

  const dx = pointerX - particle.x
  const dy = pointerY - particle.y
  const distance = Math.sqrt(dx * dx + dy * dy)
  if (distance > influenceRadius) {
    return 0
  }
  return (1 - distance / influenceRadius) ** 1.7
}

onMounted(() => {
  const canvas = canvasRef.value
  if (canvas) {
    ctx = canvas.getContext('2d')
    resizeCanvas()
    animationFrame = window.requestAnimationFrame(draw)
  }
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
  window.addEventListener('pointerleave', clearPointer)
  window.addEventListener('resize', resizeCanvas)
})

onBeforeUnmount(() => {
  window.cancelAnimationFrame(animationFrame)
  window.removeEventListener('pointermove', handlePointerMove)
  window.removeEventListener('pointerleave', clearPointer)
  window.removeEventListener('resize', resizeCanvas)
})
</script>

<template>
  <div
    class="ambient-background"
    :style="{ '--ambient-x': `${x}%`, '--ambient-y': `${y}%` }"
    aria-hidden="true"
  >
    <canvas ref="canvasRef" class="ambient-particles"></canvas>
  </div>
</template>

<style scoped>
.ambient-background {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  background:
    radial-gradient(circle at var(--ambient-x) var(--ambient-y), rgba(215, 180, 106, 0.1), transparent 24rem),
    radial-gradient(circle at 12% 20%, rgba(148, 163, 184, 0.09), transparent 22rem),
    linear-gradient(135deg, #0c0e10 0%, #141719 48%, #0a0d0e 100%);
}

.ambient-background::before {
  position: absolute;
  inset: 0;
  content: "";
  opacity: 0.28;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: radial-gradient(circle at var(--ambient-x) var(--ambient-y), black, transparent 72%);
}

.ambient-particles {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  opacity: 0.92;
  pointer-events: none;
}

@media (prefers-reduced-motion: reduce) {
  .ambient-particles {
    opacity: 0.42;
  }
}
</style>
