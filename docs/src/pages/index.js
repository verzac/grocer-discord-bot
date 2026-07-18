import React from "react";
import clsx from "clsx";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import styles from "./index.module.css";
import HomepageFeatures from "../components/HomepageFeatures";

const PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=net.grocerybot.app";
const APP_STORE_URL =
  "https://apps.apple.com/us/app/grocerybot-app/id6780291729?itscg=30200&itsct=apps_box_link&mttnsubad=6780291729";

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx("hero hero--primary", styles.heroBanner)}>
      <div className="container">
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="/docs/intro"
          >
            Get started really quickly!
          </Link>
        </div>
      </div>
    </header>
  );
}

function AppPromoSection() {
  return (
    <section className={styles.appPromo}>
      <div className="container">
        <div className={styles.appPromoInner}>
          <div className={styles.appPromoContent}>
            <h2>Now available as an app!</h2>
            <p className={styles.appPromoTagline}>
              Check your grocery lists from any Discord server in one tap.
              Lightweight, fast, and works offline.{" "}
              <Link to="/blog/new-grocerybot-app">Learn more...</Link>
            </p>
            <div className={styles.appPromoButtons}>
              <Link href={APP_STORE_URL} className={styles.storeBadge}>
                <img
                  src="/img/download-ios.svg"
                  alt="Download on the App Store"
                  height={40}
                />
              </Link>
              <Link href={PLAY_STORE_URL} className={styles.storeBadge}>
                <img
                  src="/img/download-android.svg"
                  alt="Get it on Google Play"
                  height={40}
                />
              </Link>
            </div>
            <div className={styles.storeNote}>
              <ul>
                <li>Needs your Discord account to sign-in.</li>
              </ul>
            </div>
          </div>
          <div className={styles.appScreenshot}>
            <img
              src="/img/groceryapp-screenshot.png"
              alt="GroceryBot app screenshot"
              loading="lazy"
            />
          </div>
        </div>
      </div>
    </section>
  );
}

export default function Home() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title={`Home`}
      description="Get a grocery list going for your Discord server! No sign-ups required."
    >
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <AppPromoSection />
      </main>
    </Layout>
  );
}
