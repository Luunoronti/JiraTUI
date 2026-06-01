using System.Text;
using Newtonsoft.Json.Linq;

namespace JiraTUI.Jira
{
    /// <summary>
    /// Minimal Atlassian Document Format → plain text renderer.
    /// Handles the node types you actually see in Jira issue descriptions and comments:
    /// paragraph, heading, bulletList/orderedList, listItem, codeBlock, blockquote, rule,
    /// hardBreak, text, mention, emoji, inlineCard. Unknown nodes are walked recursively.
    /// </summary>
    public static class AdfTextRenderer
    {
        public static string Render(JToken adf)
        {
            if (adf == null || adf.Type == JTokenType.Null) return "";
            var sb = new StringBuilder();
            WalkBlocks(adf, sb);
            return sb.ToString().TrimEnd();
        }

        private static void WalkBlocks(JToken node, StringBuilder sb)
        {
            var content = node["content"] as JArray;
            if (content == null) return;

            foreach (var block in content)
            {
                var type = (string)block["type"] ?? "";
                switch (type)
                {
                    case "paragraph":
                        WalkInline(block, sb);
                        sb.AppendLine();
                        break;

                    case "heading":
                        int lvl = block["attrs"]?["level"]?.Value<int>() ?? 1;
                        sb.Append(new string('#', lvl)).Append(' ');
                        WalkInline(block, sb);
                        sb.AppendLine();
                        break;

                    case "bulletList":
                        WalkList(block, sb, ordered: false);
                        break;

                    case "orderedList":
                        WalkList(block, sb, ordered: true);
                        break;

                    case "codeBlock":
                        sb.AppendLine("```");
                        WalkInline(block, sb);
                        sb.AppendLine();
                        sb.AppendLine("```");
                        break;

                    case "blockquote":
                        sb.Append("> ");
                        WalkInline(block, sb);
                        sb.AppendLine();
                        break;

                    case "rule":
                        sb.AppendLine("───");
                        break;

                    default:
                        // Unknown block — try inline content, fallback to nested blocks.
                        WalkInline(block, sb);
                        WalkBlocks(block, sb);
                        sb.AppendLine();
                        break;
                }
            }
        }

        private static void WalkInline(JToken node, StringBuilder sb)
        {
            var content = node["content"] as JArray;
            if (content == null) return;

            foreach (var n in content)
            {
                var t = (string)n["type"] ?? "";
                switch (t)
                {
                    case "text":
                        sb.Append((string)n["text"] ?? "");
                        break;

                    case "hardBreak":
                        sb.AppendLine();
                        break;

                    case "mention":
                    {
                        var mention = (string)n["attrs"]?["text"];
                        if (string.IsNullOrEmpty(mention))
                            mention = "@" + (string)n["attrs"]?["id"];
                        sb.Append(mention);
                        break;
                    }

                    case "emoji":
                        sb.Append((string)n["attrs"]?["text"]
                            ?? (string)n["attrs"]?["shortName"]
                            ?? "");
                        break;

                    case "inlineCard":
                    case "link":
                        sb.Append((string)n["attrs"]?["url"] ?? "");
                        break;

                    default:
                        WalkInline(n, sb);
                        break;
                }
            }
        }

        private static void WalkList(JToken node, StringBuilder sb, bool ordered)
        {
            var content = node["content"] as JArray;
            if (content == null) return;

            int idx = 1;
            foreach (var item in content)
            {
                sb.Append(ordered ? (idx + ". ") : "- ");
                // listItem usually wraps paragraph blocks
                var itemContent = item["content"] as JArray;
                if (itemContent != null)
                {
                    bool first = true;
                    foreach (var inner in itemContent)
                    {
                        if (!first) sb.Append("  ");
                        WalkInline(inner, sb);
                        sb.AppendLine();
                        first = false;
                    }
                }
                idx++;
            }
        }
    }
}
