import React from 'react'
import { Tab, Tabs } from 'react-bootstrap'

import { FigureData, UserData } from './types'

interface Props {
  figures: FigureData[]
  user?: UserData
}

export class Figures extends React.Component<Props> {
  public render = () => {
    const { figures, user } = this.props
    if (!figures) {
      return null
    }

    if (user) {
      return (
        <Tabs id="figures">
          <Tab title='Died on this date'>{renderFigs(figures, true)}</Tab>
          <Tab title='You have recently outlived'>
            {renderFigs(user.figures, false)}
          </Tab>
        </Tabs>
      )
    }

    return <div>{renderFigs(figures, true)}</div>
  }
}

const renderFig = (fig: FigureData, showAge: boolean) => (
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
      {showAge && <br />}
      {showAge && '(' + fig.yearsDaysAlive + ')'}
    </a>
  </div>
)

const renderFigs = (figs: FigureData[], showAge: boolean) => (
  <div>{figs.map((fig: FigureData) => renderFig(fig, showAge))}</div>
)
