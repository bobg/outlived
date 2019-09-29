export interface FigureData {
  name: string
  desc?: string
  born: string
  died: string
  daysAlive: number
  yearsDaysAlive: string
  href: string
  imgAlt?: string
  imgSrc?: string
}

export interface UserData {
  csrf: string
  born: string
  daysAlive: number
  yearsDaysAlive: string
  email: string
  figures: FigureData[]
  verified: boolean
  active: boolean
}

export interface Data {
  figures: FigureData[]
  today: string
  user?: UserData
}
