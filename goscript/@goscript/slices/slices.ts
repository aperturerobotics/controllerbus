// TypeScript implementation of Go's slices package
import * as $ from '@goscript/builtin/index.js'

/**
 * All returns an iterator over index-value pairs in the slice.
 * This is equivalent to Go's slices.All function.
 * @param s The slice to iterate over
 * @returns An iterator function that yields index-value pairs
 */
export function All<T>(
  s: $.Slice<T>,
): (yieldFunc: (index: number, value: T) => boolean) => void {
  return function (_yield: (index: number, value: T) => boolean): void {
    const length = $.len(s)
    for (let i = 0; i < length; i++) {
      const value = (s as any)[i] as T // Use proper indexing to avoid type issues
      if (!_yield(i, value)) {
        break
      }
    }
  }
}

/**
 * Sort sorts a slice in ascending order.
 * This is equivalent to Go's slices.Sort function.
 * @param s The slice to sort in place
 */
export function Sort<T extends string | number>(s: $.Slice<T>): void {
  $.sortSlice(s)
}
