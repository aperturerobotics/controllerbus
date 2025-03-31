import { Context } from '../core/core'

/**
 * DebugValues maps string key to a list of values.
 * It is used for debug visualizations.
 */
export type DebugValues = Record<string, string[]>

/**
 * Creates a new DebugValues object.
 */
export function newDebugValues(): DebugValues {
  return {}
}

/**
 * ProtoDebugValue is a protobuf-compatible debug value.
 */
export interface ProtoDebugValue {
  key: string
  values: string[]
}

/**
 * Converts DebugValues to an array of ProtoDebugValue objects.
 */
export function newProtoDebugValues(dv: DebugValues): ProtoDebugValue[] {
  const res: ProtoDebugValue[] = []
  for (const [k, v] of Object.entries(dv)) {
    res.push({
      key: k,
      values: v,
    })
  }
  return res.sort((a, b) => a.key.localeCompare(b.key))
}

/**
 * DirectiveInfo contains information about a directive.
 */
export interface DirectiveInfo {
  name: string
  debugVals?: ProtoDebugValue[]
}

/**
 * Creates a new DirectiveInfo from a directive.
 */
export function newDirectiveInfo(dir: Directive): DirectiveInfo {
  let debugVals: ProtoDebugValue[] | undefined
  if ('getDebugVals' in dir) {
    const debugDir = dir as Debuggable
    debugVals = newProtoDebugValues(debugDir.getDebugVals())
  }

  return {
    name: dir.getName(),
    debugVals,
  }
}

/**
 * DirectiveState contains information about a running directive.
 */
export interface DirectiveState {
  info: DirectiveInfo
  // TODO: add state fields
}

/**
 * Creates a new DirectiveState from a running directive.
 */
export function newDirectiveState(di: Instance): DirectiveState {
  return {
    info: newDirectiveInfo(di.getDirective()),
    // TODO: add state
  }
}

/**
 * ValueOptions are options related to value handling.
 */
export interface ValueOptions {
  /**
   * MaxValueCount indicates a maximum number of values to retrieve.
   * The resolvers will be canceled when this many values are gathered.
   * If zero, accepts infinite values.
   */
  maxValueCount: number

  /**
   * MaxValueHardCap indicates MaxValueCount is a hard cap. If it is not a
   * hard cap, any values found after resolvers are canceled is accepted. If
   * it is a hard cap, any values found after resolvers are canceled will be
   * rejected.
   */
  maxValueHardCap: boolean

  /**
   * UnrefDisposeDur is the duration to wait to dispose a directive after all
   * references have been released.
   */
  unrefDisposeDur: number

  /**
   * UnrefDisposeEmptyImmediate indicates we should immediately dispose a
   * directive that has become unreferenced if there are no associated Values
   * with the directive (it is unresolved) regardless of UnrefDisposeDur.
   */
  unrefDisposeEmptyImmediate: boolean
}

/**
 * Directive implements a requested state (with a set of values).
 */
export interface Directive {
  /**
   * Validate validates the directive.
   * This is a cursory validation to see if the values "look correct."
   */
  validate(): Error | null

  /**
   * GetValueOptions returns options relating to value handling.
   */
  getValueOptions(): ValueOptions

  /**
   * GetName returns the directives type name (i.e. DoSomething).
   * This is not intended to be unique and is primarily used for display.
   */
  getName(): string
}

/**
 * DirectiveWithEquiv contains a check to see if it is equivalent to another directive.
 */
export interface DirectiveWithEquiv extends Directive {
  /**
   * IsEquivalent checks if the other directive is equivalent. If two
   * directives are equivalent, and the new directive does not superceed the
   * old, then the new directive will be merged (de-duplicated) into the old.
   */
  isEquivalent(other: Directive): boolean
}

/**
 * DirectiveWithSuperceeds contains a check to see if the directive superceeds another.
 */
export interface DirectiveWithSuperceeds extends DirectiveWithEquiv {
  /**
   * Superceeds checks if the directive overrides another.
   * The other directive will be canceled if superceded.
   */
  superceeds(other: Directive): boolean
}

/**
 * Debuggable indicates the directive implements the DebugVals interface.
 */
export interface Debuggable {
  /**
   * GetDebugVals returns the directive arguments as key/value pairs.
   * This should be something like param1="test", param2="test".
   * This is not necessarily unique, and is primarily intended for display.
   */
  getDebugVals(): DebugValues
}

/**
 * NetworkedCodec is the encoder/decoder for a networked directive.
 */
export interface NetworkedCodec {
  /**
   * Marshal encodes the networked directive.
   */
  marshal(networked: Networked): Uint8Array

  /**
   * Unmarshal decodes the data to the networked directive.
   * The type must match the expected type for the codec.
   */
  unmarshal(data: Uint8Array, networked: Networked): void
}

/**
 * Networked is a directive which can be serialized and uniquely identified
 * across IPC domains.
 */
export interface Networked extends Directive {
  /**
   * GetNetworkedCodec returns the encoder / decoder for this directive.
   * The same encoder/decoder should also be compatible with the results.
   */
  getNetworkedCodec(): NetworkedCodec
}

/**
 * DirectiveAdder can add a directive to a bus.
 */
export interface DirectiveAdder {
  /**
   * AddDirective adds a directive to the controller.
   * This call de-duplicates equivalent directives.
   *
   * cb receives values in order as they are emitted.
   * cb can be null.
   * cb should not block.
   *
   * Returns the instance, new reference, and any error.
   */
  addDirective(
    dir: Directive,
    handler: ReferenceHandler | null,
  ): Promise<[Instance, Reference, Error | null]>
}

/**
 * HandlerAdder can add a handler to a bus.
 */
export interface HandlerAdder {
  /**
   * AddHandler adds a directive handler.
   * The handler will receive calls for all existing directives (initial set).
   * An error is returned only if adding the handler failed.
   * Returns a function to remove the handler.
   * The release function must be non-null if err is null, and null if err != null.
   */
  addHandler(handler: Handler): Promise<[() => void, Error | null]>
}

/**
 * DirectiveLister can list directives.
 */
export interface DirectiveLister {
  /**
   * GetDirectives returns a list of all currently executing directives.
   */
  getDirectives(): Instance[]
}

/**
 * Controller manages running directives and handlers.
 */
export interface Controller
  extends DirectiveLister,
    DirectiveAdder,
    HandlerAdder {}

/**
 * Reference is a reference to a directive.
 * This is used to expire directive handles.
 */
export interface Reference {
  /**
   * Release releases the reference.
   */
  release(): void
}

/**
 * ReferenceHandler handles values emitted by the directive instance.
 */
export interface ReferenceHandler {
  /**
   * HandleValueAdded is called when a value is added to the directive.
   * Should not block.
   * Avoid calling directive functions in this routine.
   */
  handleValueAdded(instance: Instance, value: AttachedValue): void

  /**
   * HandleValueRemoved is called when a value is removed from the directive.
   * Should not block.
   * Avoid calling directive functions in this routine.
   */
  handleValueRemoved(instance: Instance, value: AttachedValue): void

  /**
   * HandleInstanceDisposed is called when a directive instance is disposed.
   * This will occur if Close() is called on the directive instance.
   * Avoid calling directive functions in this routine.
   */
  handleInstanceDisposed(instance: Instance): void
}

/**
 * IdleCallback is called when the directive becomes idle or not-idle.
 * Errs is the list of non-null resolver errors.
 */
export type IdleCallback = (isIdle: boolean, errs: Error[]) => void

/**
 * Instance tracks a directive with reference counts and resolution state.
 */
export interface Instance {
  /**
   * GetContext returns a context that is canceled when Instance is released.
   */
  getContext(): Context

  /**
   * GetDirective returns the underlying directive object.
   */
  getDirective(): Directive

  /**
   * GetDirectiveIdent returns a human-readable string identifying the directive.
   *
   * Ex: DoSomething or DoSomething<param=foo>
   */
  getDirectiveIdent(): string

  /**
   * GetResolverErrors returns a snapshot of any errors returned by resolvers.
   */
  getResolverErrors(): Error[]

  /**
   * AddReference adds a reference to the directive.
   * cb is called for each value.
   * cb calls should return immediately.
   * the release callback is called immediately if already released
   * If marked as a weak ref, the handler will not count towards the ref count.
   * will never return null
   */
  addReference(cb: ReferenceHandler | null, weakRef: boolean): Reference

  /**
   * AddDisposeCallback adds a callback that will be called when the instance
   * is disposed, either when Close() is called, or when the reference count
   * drops to zero. The callback may occur immediately if the instance is
   * already disposed, but will be made in a new goroutine.
   * Returns a callback release function.
   */
  addDisposeCallback(cb: () => void): () => void

  /**
   * AddIdleCallback adds a callback that will be called when the idle state changes.
   * Called immediately with the initial state.
   * Returns a callback release function.
   */
  addIdleCallback(cb: IdleCallback): () => void

  /**
   * CloseIfUnreferenced cancels the directive instance if there are no refs.
   *
   * This bypasses the unref dispose timer.
   * If inclWeakRefs=true, keeps the instance if there are any weak refs.
   * Returns if the directive instance was closed.
   */
  closeIfUnreferenced(inclWeakRefs: boolean): boolean

  /**
   * Close cancels the directive instance and removes the directive.
   */
  close(): void
}

/**
 * Value satisfies a directive.
 */
export type Value = any

/**
 * AttachedValue is a value with some metadata.
 */
export interface AttachedValue {
  /**
   * GetValueID returns the value ID.
   */
  getValueID(): number

  /**
   * GetValue returns the value.
   */
  getValue(): Value
}

/**
 * Creates a new attached value
 */
export function newAttachedValue(vid: number, val: Value): AttachedValue {
  return new AttachedValueImpl(vid, val)
}

class AttachedValueImpl implements AttachedValue {
  constructor(
    private readonly vid: number,
    private readonly val: Value,
  ) {}

  getValueID(): number {
    return this.vid
  }

  getValue(): Value {
    return this.val
  }
}

/**
 * TypedAttachedValue is a typed value with some metadata.
 */
export interface TypedAttachedValue<T> {
  /**
   * GetValueID returns the value ID.
   */
  getValueID(): number

  /**
   * GetValue returns the value.
   */
  getValue(): T
}

/**
 * Creates a new typed attached value.
 */
export function newTypedAttachedValue<T>(
  vid: number,
  val: T,
): TypedAttachedValue<T> {
  return new TypedAttachedValueImpl<T>(vid, val)
}

class TypedAttachedValueImpl<T> implements TypedAttachedValue<T> {
  constructor(
    private readonly vid: number,
    private readonly val: T,
  ) {}

  getValueID(): number {
    return this.vid
  }

  getValue(): T {
    return this.val
  }
}

/**
 * TransformedAttachedValue is an AttachedValue with a transformed value.
 */
export interface TransformedAttachedValue<T, E> extends TypedAttachedValue<T> {
  /**
   * GetTransformedValue returns the transformed value.
   */
  getTransformedValue(): E
}

/**
 * Creates a new TransformedAttachedValue.
 */
export function newTransformedAttachedValue<T, E>(
  tav: TypedAttachedValue<T>,
  xfrm: E,
): TransformedAttachedValue<T, E> {
  return new TransformedAttachedValueImpl<T, E>(tav, xfrm)
}

class TransformedAttachedValueImpl<T, E>
  implements TransformedAttachedValue<T, E>
{
  constructor(
    private readonly tav: TypedAttachedValue<T>,
    private readonly xfrm: E,
  ) {}

  getValueID(): number {
    return this.tav.getValueID()
  }

  getValue(): T {
    return this.tav.getValue()
  }

  getTransformedValue(): E {
    return this.xfrm
  }
}

/**
 * ValueHandler handles values emitted by a resolver.
 */
export interface ValueHandler {
  /**
   * AddValue adds a value to the result, returning success and an ID. If
   * AddValue returns false, value was rejected. A rejected value should be
   * released immediately. If the value limit is reached, the value may not be
   * accepted. The value may be accepted, immediately before the resolver is
   * canceled (limit reached). It is always safe to call RemoveValue with the
   * ID at any time, even if the resolver is cancelled.
   *
   * Value IDs start at 1.
   */
  addValue(value: Value): [number, boolean]

  /**
   * RemoveValue removes a value from the result, returning found.
   * It is safe to call this function even if the resolver is canceled.
   */
  removeValue(id: number): [Value | null, boolean]

  /**
   * CountValues returns the number of values that were set.
   * if allResolvers=false, returns the number set by this handler.
   * if allResolvers=true, returns the number set by all resolvers.
   */
  countValues(allResolvers: boolean): number

  /**
   * ClearValues removes any values that were set by this handler.
   * Returns list of value IDs that were removed.
   */
  clearValues(): number[]
}

/**
 * ResolverHandler handles values emitted by the resolver and provides utils for the resolver.
 */
export interface ResolverHandler extends ValueHandler {
  /**
   * MarkIdle marks the resolver as idle or not-idle.
   * If the resolver returns null or an error, it's also marked as idle.
   */
  markIdle(idle: boolean): void

  /**
   * AddValueRemovedCallback adds a callback that will be called when the
   * given value id is disposed or removed.
   *
   * The callback will be called if the value is removed for any reason,
   * including if the parent resolver, handler, or directive are removed.
   *
   * The callback might be called immediately if the value was already removed.
   *
   * Returns a release function to clear the callback early.
   */
  addValueRemovedCallback(id: number, cb: () => void): () => void

  /**
   * AddResolverRemovedCallback adds a callback that will be called when the
   * directive resolver is removed.
   *
   * The callback will be called if the resolver is removed for any reason,
   * including if the parent resolver, handler, or directive are removed.
   *
   * The callback might be called immediately if the resolver was already removed.
   *
   * Returns a release function to clear the callback early.
   */
  addResolverRemovedCallback(cb: () => void): () => void

  /**
   * AddResolver adds a resolver as a child of the current resolver.
   *
   * The child resolver will be removed if the parent handler is removed.
   *
   * The callback will be called if the child resolver is removed for any
   * reason, including if the parent resolver, handler, or directive are
   * removed.
   *
   * The callback might be called immediately if the child resolver was
   * already removed or not created.
   *
   * Returns a release function to clear and stop the resolver early.
   * Does nothing if res == null returning an empty release func.
   */
  addResolver(res: Resolver | null, cb: () => void): () => void
}

/**
 * Resolver resolves values for directives.
 */
export interface Resolver {
  /**
   * Resolve resolves the values, emitting them to the handler.
   * The resolver may be canceled and restarted multiple times.
   * Any fatal error resolving the value is returned.
   * The resolver will not be retried after returning an error.
   * Values will be maintained from the previous call.
   */
  resolve(ctx: Context, handler: ResolverHandler): Promise<Error | null>
}

/**
 * Handler handles directives.
 */
export interface Handler {
  /**
   * HandleDirective asks if the handler can resolve the directive.
   * If it can, it returns resolver(s). If not, returns null.
   * It is safe to add a reference to the directive during this call.
   * The passed context is canceled when the directive instance expires.
   * NOTE: the passed context is not canceled when the handler is removed.
   */
  handleDirective(ctx: Context, instance: Instance): Promise<Resolver[] | null>
}

/**
 * HandlerFunc is a function that handles directives.
 */
export type HandlerFunc = (
  ctx: Context,
  instance: Instance,
) => Promise<Resolver[] | null>
