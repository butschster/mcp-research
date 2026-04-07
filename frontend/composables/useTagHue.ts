export function tagHue(tag: string): number {
  return [...tag].reduce((acc, c) => acc + c.charCodeAt(0), 0) % 6
}
