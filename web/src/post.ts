export const post = async (url: string, body: any) => {
  const resp = await fetch(url, {
    method: 'POST',
    credentials: 'same-origin',
    body: JSON.stringify(body),
    headers: { 'Content-Type': 'application/json' },
  })
  if (resp.status < 200 || resp.status >= 300) {
    throw new Error(`Server returned status ${resp.status}: ` + await resp.text())
  }
  return resp
}
