export interface FigureData {
  name: string
  desc?: string
  born: string
  died: string
  daysAlive: string
  yearsDaysAlive: string
  href: string
  imgAlt?: string
  imgSrc?: string
}

export interface UserData {
  csrf: string
  born: string
  daysAlive: string
  yearsDaysAlive: string
  email: string
  figures: FigureData[]
}

export interface Data {
  figures?: FigureData[]
  today?: string
  user?: UserData
}
