import React from 'react'
import { Tab, Tabs, TabList, TabPanel } from 'react-tabs'

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
        <Tabs>
          <TabList>
            <Tab>Died on this date</Tab>
            <Tab>Died {daysAlive - 1} days ago</Tab>
          </TabList>
          <TabPanel>{renderFigs(figures, true)}</TabPanel>
          <TabPanel>{renderFigs(user.figures, false)}</TabPanel>
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
