import React from "react";
import clsx from "clsx";
import styles from "./HomepageFeatures.module.css";

const FeatureList = [
  {
    title: "KISS: Keep It Simple, Silly",
    Svg: require("../../static/img/undraw_quitting_time.svg").default,
    description: (
      <>
        GroceryBot was designed for people who value their time. With these 2
        commands: <code>!gro</code> and <code>!grolist</code>, you can
        immediately start building your grocery list.
      </>
    ),
  },
  {
    title: "Get Organised, Fast",
    Svg: require("../../static/img/undraw_note_list.svg").default,
    description: (
      <>
        You can create multiple grocery lists on GroceryBot (
        <code>!grolist new</code>) AND have a list that automatically updates as
        you add/remove entries (<code>!grohere</code>).
      </>
    ),
  },
  {
    title: "No Sign-Ups Required",
    Svg: require("../../static/img/undraw_login.svg").default,
    description: (
      <>
        Getting started with GroceryBot is as easy as adding it to your server.
        No weird sign-up pages that takes you into an endless rabbit hole.
      </>
    ),
  },
];

function Feature({ Svg, title, description }) {
  return (
    <div className={clsx("col col--4")}>
      <div className="text--center">
        <Svg className={styles.featureSvg} alt={title} />
      </div>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
