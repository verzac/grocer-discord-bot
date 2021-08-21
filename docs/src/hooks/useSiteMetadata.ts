import { graphql, useStaticQuery } from "gatsby";

interface SiteMetadata {
  siteUrl: string;
  title: string;
}

export default function useSiteMetadata(): SiteMetadata {
  const { site } = useStaticQuery(graphql`
    query {
      site {
        siteMetadata {
          siteUrl
          title
        }
      }
    }
  `);
  return site.siteMetadata;
}
