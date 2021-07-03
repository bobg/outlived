import React, { useCallback, useState } from 'react'

import {
  AppBar,
  Box,
  Card,
  CardContent,
  Link,
  Tab,
  Typography,
} from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { TabContext, TabList, TabPanel } from '@material-ui/lab'

import { FigureData, UserData } from './types'

interface Props {
  figures: FigureData[]
  today?: string
  user: UserData | null
}

const figStyles = makeStyles({
  card: {
    display: 'inline-block',
    width: '12em',
    margin: '6px',
    textAlign: 'center',
    verticalAlign: 'top',
  },
})

export const Figures = (props: Props) => {
  const { figures, today, user } = props

  const [activeTab, setActiveTab] = useState('died-today')

  const classes = figStyles()

  if (user) {
    return (
      <Box alignItems='center' justifyContent='center' textAlign='center'>
        {today && <Typography paragraph={true}>Today is {today}.</Typography>}
        <Typography paragraph={true}>
          You were born on {user.born}, which was{' '}
          {user.daysAlive.toLocaleString()} days ago.
        </Typography>
        <TabContext value={activeTab}>
          <AppBar position='static'>
            <TabList
              onChange={(event: React.ChangeEvent<{}>, newValue: string) =>
                setActiveTab(newValue)
              }
            >
              <Tab value='died-today' label='Died on this date' />
              <Tab value='you-outlived' label='You have recently outlived' />
            </TabList>
          </AppBar>
          <TabPanel value='died-today'>
            {renderFigs(figures, true, classes)}
          </TabPanel>
          <TabPanel value='you-outlived'>
            {renderFigs(user.figures, false, classes)}
          </TabPanel>
        </TabContext>
      </Box>
    )
  }

  return (
    <div className='figures logged-out'>
      {today && <div>Today is {today}.</div>}
      <div>Died on this date:</div>
      {renderFigs(figures, true, classes)}
    </div>
  )
}

// xxx figure out the right type for classes
const renderFigs = (figs: FigureData[], showAge: boolean, classes: any) => {
  return (
    <>
      {figs.map((fig: FigureData) => (
        <Card className={classes.card}>
          <CardContent>
            <Link
              className='figure'
              target='_blank'
              rel='noopener noreferrer'
              href={fig.href}
            >
              {fig.imgSrc && (
                <span>
                  <img
                    className={classes.image}
                    src={fig.imgSrc}
                    alt={fig.imgAlt}
                  />
                  <br />
                </span>
              )}
              {fig.name}
            </Link>
            <br />
            {fig.desc}
            {fig.desc && <br />}
            {fig.born}&mdash;{fig.died}
            {showAge && <br />}
            {showAge && '(' + fig.yearsDaysAlive + ')'}
          </CardContent>
        </Card>
      ))}
    </>
  )
}
