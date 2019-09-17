import React from 'react'

import { FigureData } from './types'

interface Props {
  figures: FigureData[]
}

export class Figures extends React.Component<Props> {
  public render = () => {
    const { figures } = this.props
    if (!figures) {
      return null
    }

    console.log(`rendering Figures, figures is ${JSON.stringify(figures)}`)

    return (
      <div>
        {figures.map((fig: any) => (
          <div>
            <a
              className='figure'
              target='_blank'
              rel='noopener noreferrer'
              href={fig.href}
            >
              {fig.imgSrc && (
                <img className='img64' src={fig.imgSrc} alt={fig.imgAlt} />
              )}
              {fig.name}
              <br />
              {fig.desc}
              {fig.desc && <br />}
              {fig.born}&mdash;{fig.died}
            </a>
          </div>
        ))}
      </div>
    )
  }
}
