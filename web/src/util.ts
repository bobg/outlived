export const datestr = (d: Date) => {
  const y = d.getUTCFullYear()
  const m = 1 + d.getUTCMonth() // come on javascript
  const day = d.getUTCDate() // come ON, javascript
  return y + '-' + ('0' + m).substr(-2) + '-' + ('0' + day).substr(-2)
}

export const daysInMonth = (y: number, m: number) => {
  switch (m) {
    case 1:
    case 3:
    case 5:
    case 7:
    case 8:
    case 10:
    case 12:
      return 31

    case 4:
    case 6:
    case 9:
    case 11:
      return 30
  }

  if (y % 400 === 0) {
    return 29
  }
  if (y % 100 === 0) {
    return 28
  }
  if (y % 4 === 0) {
    return 29
  }
  return 28
}
