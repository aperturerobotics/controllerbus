import { describe, it, expect, vi, beforeEach } from 'vitest'
import { DirectiveInstance, DirectiveControllerImpl, newController } from './controller'
import { background } from '../../core/core'
import { 
  Directive, 
  ValueOptions, 
  Handler,
  Resolver,
  ResolverHandler,
  DirectiveWithEquiv,
  DirectiveWithSuperceeds
} from '../directive'

// Mock logger implementation
class MockLogger {
  debug = vi.fn()
  info = vi.fn()
  warn = vi.fn()
  error = vi.fn()
  debugf = vi.fn()
  infof = vi.fn()
  warnf = vi.fn()
  errorf = vi.fn()
}

// Basic directive implementation
class TestDirective implements Directive {
  constructor(private name: string = 'TestDirective') {}
  
  validate(): Error | null {
    return null
  }
  
  getValueOptions(): ValueOptions {
    return {
      maxValueCount: 0,
      maxValueHardCap: false,
      unrefDisposeDur: 0,
      unrefDisposeEmptyImmediate: true
    }
  }
  
  getName(): string {
    return this.name
  }
}

// Directive with equivalence checks
class EquivDirective implements DirectiveWithEquiv {
  constructor(private id: string) {}
  
  validate(): Error | null {
    return null
  }
  
  getValueOptions(): ValueOptions {
    return {
      maxValueCount: 0,
      maxValueHardCap: false,
      unrefDisposeDur: 0,
      unrefDisposeEmptyImmediate: true
    }
  }
  
  getName(): string {
    return 'EquivDirective'
  }
  
  isEquivalent(other: Directive): boolean {
    if (!(other instanceof EquivDirective)) {
      return false
    }
    return this.id === (other as EquivDirective).id
  }
}

// Directive with superceeds
class SupercedesDirective extends EquivDirective implements DirectiveWithSuperceeds {
  constructor(id: string, private version: number) {
    super(id)
  }
  
  superceeds(other: Directive): boolean {
    if (!(other instanceof SupercedesDirective)) {
      return false
    }
    return this.version > (other as SupercedesDirective).version
  }
}

describe('DirectiveController', () => {
  let logger: MockLogger
  let ctx: any
  
  beforeEach(() => {
    logger = new MockLogger()
    ctx = background()
    
    // Add context to global for DirectiveInstance to use
    ;(globalThis as any).context = {
      withCancel: () => ctx.withCancel()
    }
  })
  
  describe('DirectiveInstance', () => {
    it('should create a directive instance with correct identifier', () => {
      const directive = new TestDirective('TestDir')
      const instance = new DirectiveInstance(directive, logger)
      
      expect(instance.getDirective()).toBe(directive)
      expect(instance.getDirectiveIdent()).toBe('TestDir')
      expect(instance.getResolverErrors()).toEqual([])
    })
    
    it('should handle reference counting', () => {
      const directive = new TestDirective()
      const instance = new DirectiveInstance(directive, logger)
      
      const mockHandler = {
        handleValueAdded: vi.fn(),
        handleValueRemoved: vi.fn(),
        handleInstanceDisposed: vi.fn()
      }
      
      // Add a reference
      const ref = instance.addReference(mockHandler, false)
      expect(ref).toBeDefined()
      
      // Close and verify handler was called
      instance.close()
      expect(mockHandler.handleInstanceDisposed).toHaveBeenCalledWith(instance)
    })
    
    it('should handle dispose callbacks', () => {
      const directive = new TestDirective()
      const instance = new DirectiveInstance(directive, logger)
      
      const disposeCallback = vi.fn()
      const release = instance.addDisposeCallback(disposeCallback)
      
      expect(typeof release).toBe('function')
      
      // Close and verify callback was called
      instance.close()
      expect(disposeCallback).toHaveBeenCalled()
    })
    
    it('should handle idle callbacks', () => {
      const directive = new TestDirective()
      const instance = new DirectiveInstance(directive, logger)
      
      const idleCallback = vi.fn()
      const release = instance.addIdleCallback(idleCallback)
      
      // Should be called immediately with initial state
      expect(idleCallback).toHaveBeenCalledWith(true, [])
      
      // Update idle state and verify callback was called again
      instance.updateIdleState(false)
      expect(idleCallback).toHaveBeenCalledWith(false, [])
    })
    
    it('should close if unreferenced when configured to do so', () => {
      const directive = new TestDirective()
      directive.getValueOptions = () => ({
        maxValueCount: 0,
        maxValueHardCap: false,
        unrefDisposeDur: 0,
        unrefDisposeEmptyImmediate: true
      })
      
      const instance = new DirectiveInstance(directive, logger)
      
      // Should be closed because no references
      const closed = instance.closeIfUnreferenced(false)
      expect(closed).toBe(true)
    })
  })
  
  describe('DirectiveControllerImpl', () => {
    it('should create a controller with newController factory', () => {
      const controller = newController(ctx, logger)
      expect(controller).toBeDefined()
      expect(controller.getDirectives).toBeDefined()
      expect(controller.addDirective).toBeDefined()
      expect(controller.addHandler).toBeDefined()
    })
    
    it('should add and get directives', async () => {
      const controller = new DirectiveControllerImpl(ctx, logger)
      const directive = new TestDirective()
      
      const [instance, ref, err] = await controller.addDirective(directive, null)
      
      expect(err).toBeNull()
      expect(instance).toBeDefined()
      expect(ref).toBeDefined()
      
      const directives = controller.getDirectives()
      expect(directives).toHaveLength(1)
      expect(directives[0]).toBe(instance)
    })
    
    it('should handle validation errors', async () => {
      const controller = new DirectiveControllerImpl(ctx, logger)
      const directive = new TestDirective()
      directive.validate = () => new Error('validation error')
      
      const [instance, ref, err] = await controller.addDirective(directive, null)
      
      expect(err).not.toBeNull()
      expect(err?.message).toBe('validation error')
      expect(instance).toBeNull()
      expect(ref).toBeNull()
    })
    
    it('should handle equivalent directives', async () => {
      const controller = new DirectiveControllerImpl(ctx, logger)
      const dir1 = new EquivDirective('same-id')
      const dir2 = new EquivDirective('same-id')
      
      // Add first directive
      const [instance1, ref1, err1] = await controller.addDirective(dir1, null)
      expect(err1).toBeNull()
      
      // Add equivalent directive - should return the same instance
      const [instance2, ref2, err2] = await controller.addDirective(dir2, null)
      expect(err2).toBeNull()
      expect(instance2).toBe(instance1)
      expect(ref2).not.toBe(ref1)
      
      // Should only have one directive
      const directives = controller.getDirectives()
      expect(directives).toHaveLength(1)
    })
    
    it('should handle superceeding directives', async () => {
      const controller = new DirectiveControllerImpl(ctx, logger)
      const dir1 = new SupercedesDirective('same-id', 1)
      const dir2 = new SupercedesDirective('same-id', 2)
      
      // Add first directive
      const [instance1, ref1, err1] = await controller.addDirective(dir1, null)
      expect(err1).toBeNull()
      
      // Add superceeding directive - should replace the first one
      const [instance2, ref2, err2] = await controller.addDirective(dir2, null)
      expect(err2).toBeNull()
      expect(instance2).not.toBe(instance1)
      
      // Should only have one directive (the newer one)
      const directives = controller.getDirectives()
      expect(directives).toHaveLength(1)
      expect(directives[0]).toBe(instance2)
    })
    
    it('should add handlers and notify them of existing directives', async () => {
      const controller = new DirectiveControllerImpl(ctx, logger)
      const directive = new TestDirective()
      
      // Add a directive first
      const [instance, ref, err] = await controller.addDirective(directive, null)
      expect(err).toBeNull()
      
      // Create a handler that will be notified of the existing directive
      const handleDirective = vi.fn().mockResolvedValue(null)
      const handler: Handler = {
        handleDirective
      }
      
      // Add the handler
      const [release, handlerErr] = await controller.addHandler(handler)
      expect(handlerErr).toBeNull()
      expect(typeof release).toBe('function')
      
      // Handler should be notified of existing directive
      expect(handleDirective).toHaveBeenCalledWith(instance.getContext(), instance)
      
      // Add another directive and verify handler is called
      handleDirective.mockClear()
      const dir2 = new TestDirective('Another')
      await controller.addDirective(dir2, null)
      
      expect(handleDirective).toHaveBeenCalled()
      
      // Remove handler
      release()
      
      // Add one more directive - handler should not be called
      handleDirective.mockClear()
      const dir3 = new TestDirective('YetAnother')
      await controller.addDirective(dir3, null)
      
      expect(handleDirective).not.toHaveBeenCalled()
    })
  })
})