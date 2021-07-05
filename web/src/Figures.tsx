import React, { useCallback, useState } from 'react'

import {
  AppBar,
  Box,
  Card,
  CardContent,
  Link,
  Paper,
  Tab,
  Typography,
} from '@material-ui/core'
import { makeStyles, Theme, useTheme } from '@material-ui/core/styles'
import { TabContext, TabList, TabPanel } from '@material-ui/lab'

import { FigureData, UserData } from './types'

interface Props {
  diedToday: FigureData[]
  outlived: FigureData[] | null
}

const useStyles = (theme: Theme) =>
  makeStyles({
    card: {
      display: 'inline-block',
      width: '14em',
      margin: '6px',
      textAlign: 'center',
      verticalAlign: 'top',
    },
  })

export const Figures = (props: Props) => {
  const { diedToday, outlived } = props

  const [activeTab, setActiveTab] = useState('died-today')

  const theme = useTheme()
  const classes = useStyles(theme)()

  return (
    <Box alignItems='center' justifyContent='center' textAlign='center'>
      {outlived ? (
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
            {renderFigs(diedToday, true, classes)}
          </TabPanel>
          <TabPanel value='you-outlived'>
            {renderFigs(outlived, false, classes)}
          </TabPanel>
        </TabContext>
      ) : (
        <>
          <Typography paragraph={true}>Died on this date:</Typography>
          {renderFigs(diedToday, true, classes)}
        </>
      )}
    </Box>
  )
}

// xxx figure out the right type for classes
const renderFigs = (figs: FigureData[], showAge: boolean, classes: any) => {
  showAge = true // xxx

  return (
    <>
      {figs.map((fig: FigureData) => (
        <Card className={classes.card} key={fig.href}>
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
