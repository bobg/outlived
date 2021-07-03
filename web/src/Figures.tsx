import React from 'react'
import { AppBar, Card, Grid, Tab, Tabs } from '@material-ui/core'
import { TabPanel } from '@material-ui/lab'

import { FigureData, UserData } from './types'

interface Props {
  figures: FigureData[]
  today?: string
  user: UserData | null
}

interface State {
  activeTab: any
}

export class Figures extends React.Component<Props, State> {
  public state: State = { activeTab: '2' }

  private handleTab = (ev: React.ChangeEvent<{}>, value: any) => {
    this.setState({ activeTab: value })
  }

  public render = () => {
    const { figures, today, user } = this.props
    if (!figures) {
      return null
    }

    if (user) {
      const { activeTab } = this.state

      return (
        <div className='figures logged-in'>
          {today && <div>Today is {today}.</div>}
          <div>
            You were born on {user.born}, which was{' '}
            {user.daysAlive.toLocaleString()} days ago.
          </div>
          <AppBar position='static'>
            <Tabs id='figures' value={activeTab} onChange={this.handleTab}>
              <Tab label='Died on this date' />
              <Tab label='You have recently outlived' />
            </Tabs>
          </AppBar>
          <TabPanel value={'0'}>{renderFigs(figures, true)}</TabPanel>
          <TabPanel value={'1'}>{renderFigs(user.figures, false)}</TabPanel>
        </div>
      )
    }

    return (
      <div className='figures logged-out'>
        {today && <div>Today is {today}.</div>}
        <div>Died on this date:</div>
        {renderFigs(figures, true)}
      </div>
    )
  }
}

const renderFigs = (figs: FigureData[], showAge: boolean) => (
  <Grid>
    {figs.map((fig: FigureData) => (
      <Card variant='outlined'>
        <a
          className='figure'
          target='_blank'
          rel='noopener noreferrer'
          href={fig.href}
        >
          {fig.imgSrc && (
            <span>
              <img className='img128' src={fig.imgSrc} alt={fig.imgAlt} />
              <br />
            </span>
          )}
          {fig.name}
        </a>
        <br />
        {fig.desc}
        {fig.desc && <br />}
        {fig.born}&mdash;{fig.died}
        {showAge && <br />}
        {showAge && '(' + fig.yearsDaysAlive + ')'}
      </Card>
    ))}
  </Grid>
)
