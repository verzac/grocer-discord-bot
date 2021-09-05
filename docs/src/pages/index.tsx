import React from "react";
import {
  Box,
  Button,
  Card,
  CardHeader,
  Divider,
  Grid,
  makeStyles,
  Typography,
  Link as MuiLink,
} from "@material-ui/core";
import HeroContainer from "components/HeroContainer";
import styled from "styled-components";
import PageContainer from "components/PageContainer";
import ArrowDownwardIcon from "@material-ui/icons/ArrowDownward";
import { StaticImage } from "gatsby-plugin-image";
import DiscordLogo from "images/discord-logo.svg";
import CommandText from "components/CommandText";
import { OutboundLink } from "gatsby-plugin-google-gtag";

const CtaContainer = styled.div`
  display: flex;
  flex-direction: row;
  & > * {
    margin: 8px;
  }
`;

const FeatureCard = styled(Card)`
  height: 100%;
  box-shadow: 0 0 14px rgba(0, 0, 0, 0.5);
  /* display: flex;
  flex-direction: column;
  align-items: flex-end; */
`;

const FeatureCardHeader = styled(CardHeader)``;

const useStyles = makeStyles((theme) => ({
  logo: {
    // maxWidth: 256,
    borderRadius: 90,
    backgroundColor: theme.palette.secondary.main,
  },
  addToDiscordBtn: {
    // maxHeight: 32,
    width: "auto",
    backgroundColor: "#5865F2",
    "&:hover": {
      backgroundColor: "#5865F2",
    },
    "& > .MuiButton-label": {
      color: "#ffffff",
      fontWeight: 700,
      display: "flex",
      flexDirection: "column",
      alignItems: "center",
    },
  },
  featureImage: {
    boxShadow: "0 0 8px rgba(0, 0, 0, 0.5)",
    objectFit: "contain",
    marginTop: "auto",
  },
}));

const FeatureCardSubText = styled.span`
  display: block;
  line-height: 1.5;
  font-size: 0.75rem;
  padding-top: 2px;
`;

function IndexPage() {
  const classes = useStyles();
  return (
    <PageContainer
      rootContainerProps={{ noDefaultFlex: true }}
      subtitle="Manage your grocery list on Discord"
    >
      <>
        <HeroContainer>
          <StaticImage
            src="../images/grobot-logo.png"
            className={classes.logo}
            alt="GroceryBot Logo"
            aspectRatio={1}
            width={256}
          />
          <Typography variant="h1" align="center">
            GroceryBot
          </Typography>
          <Typography variant="h2" align="center">
            Get a grocery list going for your Discord server!
          </Typography>
          <CtaContainer>
            <Button
              className={classes.addToDiscordBtn}
              component={OutboundLink}
              href="https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot"
              target="_blank"
              rel="noopener"
              variant="contained"
            >
              Add to Discord
              <img
                src={DiscordLogo}
                style={{
                  height: "32px",
                  width: "auto",
                }}
                alt="discord logo"
              />
            </Button>
          </CtaContainer>
          <Typography align="center">
            Need help? Join our{" "}
            <MuiLink
              href="https://discord.com/invite/rBjUaZyskg"
              rel="noopener"
              target="_blank"
            >
              support server
            </MuiLink>{" "}
            here!
          </Typography>
          <Button
            variant="text"
            color="primary"
            onClick={() => {
              const pos = document
                .getElementById("see-more")
                ?.getBoundingClientRect().top;
              if (!pos) {
                console.error("Cannot scroll: pos not found.");
                return;
              }
              window.scrollTo({
                top: pos - 64,
                behavior: "smooth",
              });
            }}
          >
            <Box display="flex" flexDirection="column" alignItems="center">
              See more
              <ArrowDownwardIcon />
            </Box>
          </Button>
        </HeroContainer>
        <Divider />
        <section id="see-more">
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} lg={3}>
              <FeatureCard>
                <StaticImage
                  src="../images/feature-gro.png"
                  alt="example for adding a grocery"
                  className={classes.featureImage}
                />
                <FeatureCardHeader
                  title="Add things you want to buy in the middle of a conversation"
                  subheader={
                    <>
                      <span>
                        Accidentally thought about that awesome thing you and
                        your housemates should get while hanging out in Discord?
                        Just type <CommandText>!gro</CommandText> to enter it
                        into your grocery list, so that <s>you</s> Kyle can buy
                        it next time!
                      </span>
                      <Divider />
                      <FeatureCardSubText>
                        Sorry Kyle - I think you need better friends.
                      </FeatureCardSubText>
                    </>
                  }
                />
              </FeatureCard>
            </Grid>
            <Grid item xs={12} sm={6} lg={3}>
              <FeatureCard>
                <StaticImage
                  src="../images/feature-grohere2.png"
                  alt="example of a dynamic grocery list"
                  className={classes.featureImage}
                />
                <FeatureCardHeader
                  title="Dynamic grocery list"
                  subheader={
                    <>
                      <span>
                        Use <CommandText>!grohere</CommandText> in an empty
                        channel and let GroceryBot always display your latest
                        grocery list as you update them. No more typing{" "}
                        <CommandText>!grolist</CommandText> while you&apos;re
                        holding ten cans of tuna - just switch channels to look
                        at your grocery list.
                      </span>
                      <Divider />
                      <FeatureCardSubText>
                        Steve Jobs hated the fact that early computer users had
                        to type to get their computer to do anything - so do we.
                      </FeatureCardSubText>
                    </>
                  }
                />
              </FeatureCard>
            </Grid>
            <Grid item xs={12} sm={6} lg={3}>
              <FeatureCard>
                <StaticImage
                  src="../images/feature-webhook.png"
                  alt="example for using webhook with GroceryBot"
                  className={classes.featureImage}
                />
                <FeatureCardHeader
                  title="Automation friendly"
                  subheader={
                    <>
                      <span>
                        Are you an aspiring developer/hacker? GroceryBot
                        responds to bot commands and webhooks, so you can
                        automate it to your heart&apos;s content!
                      </span>
                      <Divider />
                      <FeatureCardSubText>
                        (PS: the team connected someone&apos;s Google Assistant
                        to GroceryBot - how cool is that!?)
                      </FeatureCardSubText>
                    </>
                  }
                />
              </FeatureCard>
            </Grid>
            <Grid item xs={12} sm={6} lg={3}>
              <FeatureCard>
                <StaticImage
                  src="../images/feature-groreset.png"
                  alt="example of removing your data from GroceryBot"
                  className={classes.featureImage}
                />
                <FeatureCardHeader
                  title="Privacy first"
                  subheader={
                    <>
                      <span>
                        We do NOT share your data with any third-parties, nor do
                        we use your data to &quot;improve&quot; our services.
                        Don&apos;t like GroceryBot? Just type{" "}
                        <CommandText>!groreset</CommandText> to remove all of
                        your data.
                      </span>
                      <Divider />
                      <FeatureCardSubText>
                        Did you expect a light-hearted comment here? Sorry, but
                        we really take the privacy of our users seriously, so no
                        funny quips here.
                      </FeatureCardSubText>
                    </>
                  }
                />
              </FeatureCard>
            </Grid>
          </Grid>
        </section>
      </>
    </PageContainer>
  );
}

export default IndexPage;
