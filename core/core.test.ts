import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { background } from './core'

describe('Context', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('background context is never done', async () => {
    const ctx = background()
    expect(ctx.isDone()).toBe(false)

    // Set up a promise that resolves when done
    const donePromise = Promise.race([
      ctx.done(),
      new Promise((resolve) => setTimeout(() => resolve('timeout'), 100)),
    ])

    vi.advanceTimersByTime(100)
    const result = await donePromise
    expect(result).toBe('timeout')
  })

  it('withCancel creates cancellable context', async () => {
    const ctx = background()
    const [cancelCtx, cancelFn] = ctx.withCancel()

    expect(cancelCtx.isDone()).toBe(false)

    const donePromise = cancelCtx.done()
    cancelFn()

    expect(cancelCtx.isDone()).toBe(true)
    await expect(donePromise).resolves.toBeUndefined()
  })

  it('withTimeout creates auto-cancelling context', async () => {
    const ctx = background()
    const [timeoutCtx, cancelFn] = ctx.withTimeout(500)

    expect(timeoutCtx.isDone()).toBe(false)

    const donePromise = timeoutCtx.done()
    vi.advanceTimersByTime(500)

    expect(timeoutCtx.isDone()).toBe(true)
    await expect(donePromise).resolves.toBeUndefined()
  })

  it('withTimeout can be cancelled manually before timeout', async () => {
    const ctx = background()
    const [timeoutCtx, cancelFn] = ctx.withTimeout(500)

    expect(timeoutCtx.isDone()).toBe(false)

    const donePromise = timeoutCtx.done()
    cancelFn()

    expect(timeoutCtx.isDone()).toBe(true)
    await expect(donePromise).resolves.toBeUndefined()

    // Advance time past timeout to ensure we don't double-cancel
    vi.advanceTimersByTime(600)
    expect(timeoutCtx.isDone()).toBe(true)
  })

  it('child context is cancelled when parent is cancelled', async () => {
    const ctx = background()
    const [parentCtx, parentCancelFn] = ctx.withCancel()
    const [childCtx, childCancelFn] = parentCtx.withCancel()

    expect(parentCtx.isDone()).toBe(false)
    expect(childCtx.isDone()).toBe(false)

    const childDonePromise = childCtx.done()
    parentCancelFn()

    expect(parentCtx.isDone()).toBe(true)
    expect(childCtx.isDone()).toBe(true)
    await expect(childDonePromise).resolves.toBeUndefined()
  })
})
