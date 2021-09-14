import {
  Grid,
  List,
  ListItem,
  ListItemText,
  makeStyles,
  Theme,
  Typography,
  withTheme,
} from "@material-ui/core";
import CommandCard from "components/docs/CommandCard";
import PageContainer from "components/PageContainer";
import React from "react";
import styled from "styled-components";

const ContainerGrid = styled(Grid)`
  width: 100%;
`;

const DocsPageContainer = withTheme(styled(PageContainer)`
  ${({ theme }: { theme: Theme }) => theme.breakpoints.up("sm")} {
    padding-left: 0;
  }
`);

const useStyles = makeStyles((theme: Theme) => ({
  navBarContainer: {
    [theme.breakpoints.up("sm")]: {
      borderRight: `1px solid ${theme.palette.divider}`,
    },
    minHeight: "100%",
  },
  commandCard: {
    height: "100%",
  },
}));

function DocsPage() {
  const classes = useStyles();
  return (
    <DocsPageContainer subtitle="Docs">
      <ContainerGrid container spacing={1}>
        <Grid item xs={12} sm={3}>
          <List
            component="nav"
            aria-label="documentation navbar"
            className={classes.navBarContainer}
          >
            <ListItem button>
              <ListItemText primary="Introduction" />
            </ListItem>
          </List>
        </Grid>
        <Grid item xs={12} sm={9}>
          <Typography variant="h1">Docs</Typography>
          <section>
            <Typography variant="h2">Introduction</Typography>
            <Typography gutterBottom>
              GroceryBot allows you to have a super-charged grocery list right
              in Discord, where you talk with your housemates and friends. When
              you go do your grocery, we dare you to count how many times you
              switch apps to (1) talk with your mates; and (2) look at your
              grocery list.
            </Typography>
            <Typography gutterBottom>
              Switching apps is&apos;t a good experience.
            </Typography>
            <Typography gutterBottom>
              That&apos;s why we made GroceryBot. We want to allow you to have
              something as simple as a grocery list be integrated with your
              daily conversations. And we want to do it in the most
              non-intrusive way possible.
            </Typography>
          </section>
          <section>
            <Typography variant="h2">Basic Commands</Typography>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <CommandCard
                  className={classes.commandCard}
                  command="!grohelp"
                  description="Shows all the available commands that GroceryBot will respond to (and a bunch of other useful tips - basically all you need to know about GroceryBot)."
                />
              </Grid>
              <Grid item xs={12}>
                <CommandCard
                  className={classes.commandCard}
                  command="!gro"
                  description="Adds an item to your grocery list."
                />
              </Grid>
            </Grid>
          </section>
        </Grid>
      </ContainerGrid>
    </DocsPageContainer>
  );
}

export default DocsPage;
