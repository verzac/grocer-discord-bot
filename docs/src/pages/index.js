import React from "react";
import clsx from "clsx";
import Layout from "@theme/Layout";
import Link from "@docusaurus/Link";
import useDocusaurusContext from "@docusaurus/useDocusaurusContext";
import styles from "./index.module.css";
import HomepageFeatures from "../components/HomepageFeatures";

const PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=net.grocerybot.app";

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
              Lightweight, fast, and works offline.
            </p>
            <div className={styles.appPromoButtons}>
              <Link
                className="button button--primary button--lg"
                href={PLAY_STORE_URL}
              >
                Download on Google Play
              </Link>
              <Link
                className="button button--outline button--primary button--lg"
                to="/blog/new-grocerybot-app"
              >
                Learn more
              </Link>
            </div>
            <div className={styles.iosNote}>
              <ul>
                <li>
                  On iOS? Use <code>/waitlist ios</code> in Discord to get
                  notified when it launches.
                </li>
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
