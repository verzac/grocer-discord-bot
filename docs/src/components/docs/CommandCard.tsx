import {
  Card,
  CardHeader,
  Grid,
  makeStyles,
  Typography,
  WithTheme,
} from "@material-ui/core";
import clsx from "clsx";
import { StaticImage } from "gatsby-plugin-image";
import React from "react";
import styled from "styled-components";

interface CommandCardProps {
  command: string;
  description: string;
  // image: React.ReactNode;
  className?: string;
}

const useStyles = makeStyles((theme) => ({
  description: {
    display: "flex",
    flexDirection: "column",
    justifyContent: "center",
  },
  container: {
    padding: theme.spacing(2),
    marginTop: "auto",
    marginBottom: "auto",
  },
  heroImg: {
    // height: "100%",
    // marginTop: "auto",
    // marginBottom: "auto",
  },
  card: {
    // marginTop: "auto",
    // marginBottom: "auto",
  },
  sampleCommand: {
    padding: theme.spacing(1),
    fontFamily: "Consolas, Monaco, 'Andale Mono', 'Ubuntu Mono', monospace",
  },
  sampleCommandContainer: {
    display: "block",
    background: "gray",
  },
}));

function CommandCard({ command, description, className }: CommandCardProps) {
  const classes = useStyles();
  return (
    <Card className={clsx(className, classes.card)}>
      <Grid container spacing={1} className={classes.container}>
        <Grid className={classes.description} item xs={12} md={12}>
          <Typography variant="h6">{command}</Typography>
          <Typography>{description}</Typography>
        </Grid>
        {/* <Grid item xs={12} md={12} className={classes.heroImg}> */}
        {/* <StaticImage
            // layout="fullWidth"
            src="../images/feature-gro.png"
            alt="example for adding a grocery"
          /> */}
        {/* </Grid> */}
        <Grid item xs={12}>
          <Typography>Sample usage:</Typography>
          <div className={classes.sampleCommandContainer}>
            <code className={classes.sampleCommand}>!gro chicken</code>
          </div>
        </Grid>
      </Grid>
    </Card>
  );
}

export default CommandCard;
