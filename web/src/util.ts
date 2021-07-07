export const datestr = (d: Date) => {
  const y = d.getUTCFullYear()
  const m = 1 + d.getUTCMonth() // come on javascript
  const day = d.getUTCDate() // come ON, javascript
  return y + '-' + ('0' + m).substr(-2) + '-' + ('0' + day).substr(-2)
}
