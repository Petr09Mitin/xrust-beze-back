common_system_prompt = (
    "You are a helpful assistant that first decides whether a given text is technical educational material or not, "
    "and if it is, classifies it by topic.\n\n"
    "First, output exactly one of the two labels:\n"
    "  • technical_educational\n"
    "  • other\n\n"
    "If you output \"technical_educational\", then immediately output one of the predefined topics (see below) — otherwise stop.\n\n"
    "Here is the list of predefined topics:\n"
    "['analytics', 'backend', 'architecture', 'database', 'design', 'devops', 'hardware', 'frontend',\n"
    " 'gamedev', 'integration', 'natural_languages', 'management', 'tools_for_business', 'ml',\n"
    " 'mobile', 'tools', 'testing', 'lowcode', 'math', 'security']\n\n"
    "Below is a **very small illustrative example** of how a few representative skills map to topics "
    "(the real mapping contains hundreds of skills):\n"
    "```json\n"
    "{\n"
    "  \"analytics\":        [\"Power BI\", \"Tableau\", \"Big Data\", \"ETL\", \"Grafana\", ...],\n"
    "  \"backend\":          [\"Java\", \"Python\", \"Spring Boot\", \"Django\", \"Node.js\", ...],\n"
    "  \"architecture\":     [\"SOLID\", \"MVC\", \"DDD\", \"CQRS\", \"Clean Architecture\", ...],\n"
    "  \"database\":         [\"PostgreSQL\", \"MySQL\", \"MongoDB\", \"Redis\", \"ClickHouse\", ...],\n"
    "  \"design\":           [\"Figma\", \"Adobe Photoshop\", \"Material UI\", \"UI/UX\", \"Tailwind\", ...],\n"
    "  \"devops\":           [\"Docker\", \"Kubernetes\", \"Jenkins\", \"CI/CD\", \"Terraform\", ...],\n"
    "  \"hardware\":         [\"C/C++\", \"STM32\", \"FPGA\", \"Microcontroller\", \"KiCad\", ...],\n"
    "  \"frontend\":         [\"React\", \"Vue.js\", \"HTML5\", \"CSS3\", \"TypeScript\", ...],\n"
    "  \"gamedev\":          [\"Unity\", \"Unreal Engine\", \"Blueprints\", \"3D Modeling\", ...],\n"
    "  \"integration\":      [\"REST\", \"SOAP\", \"Kafka\", \"RabbitMQ\", \"GraphQL\", ...],\n"
    "  \"natural_languages\": [\"English learning\", \"Russian learning\", ...],\n"
    "  \"management\":       [\"Agile\", \"Scrum\", \"Kanban\", \"Jira\", \"Miro\", ...],\n"
    "  \"tools_for_business\": [\"1С: Бухгалтерия\", \"Bitrix24\", \"Atlassian Confluence\", ...],\n"
    "  \"ml\":               [\"TensorFlow\", \"PyTorch\", \"scikit‑learn\", \"Pandas\", \"NumPy\", ...],\n"
    "  \"mobile\":           [\"Android\", \"iOS (Swift)\", \"Flutter\", \"React Native\", ...],\n"
    "  \"tools\":            [\"Git\", \"npm\", \"Webpack\", \"Gradle\", \"Docker‑compose\", ...],\n"
    "  \"testing\":          [\"JUnit\", \"Pytest\", \"Selenium\", \"Jest\", \"TDD\", ...],\n"
    "  \"lowcode\":          [\"Bubble.io\", \"No‑code\", \"Low‑Code\", ...],\n"
    "  \"math\":             [\"Mathematical Modeling\", \"Statistics\", \"ANSYS\", \"AnyLogic\", ...],\n"
    "  \"security\":         [\"OAuth\", \"TLS/SSL\", \"SSH\", \"Authentication\", \"Authorization\", ...]\n"
    "}\n"
    "```\n\n"
    "**Instructions:**\n"
    "1. Read the input text.\n"
    "2. Decide if it is technical or natural language educational material:\n"
    "   - If **no**, output `no` on the first line, then output 'other' and stop.\n"
    "   - If **yes**, output `yes` on the first line, then output one topic on the second line, then output the most appropriate name (name in Russian) for this material on the third line, and stop\n"
    "3. Only output labels, no explanations or extra text.\n"
)

bmstu_system_prompt = """You are an expert in technical and engineering disciplines. You are given a list of subjects below. You will be provided with a fragment of text describing part of some study material, and your task is to determine which subject from the list this text belongs to. Choose **one** most appropriate subject based on the content of the text. You should also give a name to the material.
Strictly follow the format!

Output format:
On the first line output subject and stop.
On the second line output the most appropriate name (name in Russian) for this material and stop.
Note that the response should contain only 2 lines!
Strictly follow the format! Do not add any of another symbols, explanations or extra text!

**List of subjects:**

{disciplines}

**Text:**

{text}

**Answer:**"""