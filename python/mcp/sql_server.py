import json
import psycopg2
import csv
from mcp.server.fastmcp import FastMCP

# 初始化 MCP 服务器
mcp = FastMCP("SQLServer")
USER_AGENT = "SQLserver-app/1.0"


@mcp.tool()
async def sql_inter(sql_query):
    """
    查询本地Postgres数据库，通过运行一段SQL代码来进行数据库查询。\
    :param sql_query: 字符串形式的SQL查询语句
    :return：sql_query在MySQL中的运行结果。
    """

    connection = psycopg2.connect(
        host='localhost',
        user='root',
        passwd='123456',
        db='postgres',
    )

    try:
        with connection.cursor() as cursor:
            sql = sql_query
            cursor.execute(sql)
            results = cursor.fetchall()

    finally:
        connection.close()

    return json.dumps(results)


@mcp.tool()
async def export_table_to_csv(table_name, output_file):
    """
    将 Postgres 数据库中的某个表导出为 CSV 文件。

    :param table_name: 需要导出的表名
    :param output_file: 输出的 CSV 文件路径
    """
    connection = psycopg2.connect(
        host='localhost',  # 数据库地址
        user='root',  # 数据库用户名
        passwd='123',  # 数据库密码
        db='school',  # 数据库名
        charset='utf8'  # 字符集
    )

    try:
        with connection.cursor() as cursor:
            query = f"SELECT * FROM {table_name};"
            cursor.execute(query)

            column_names = [desc[0] for desc in cursor.description]

            rows = cursor.fetchall()

            with open(output_file, mode='w', newline='', encoding='utf-8') as file:
                writer = csv.writer(file)

                writer.writerow(column_names)

                writer.writerows(rows)

            print(f"数据表 {table_name} 已成功导出至 {output_file}")

    except Exception as e:
        print(f"导出失败: {e}")

    finally:
        connection.close()


if __name__ == "__main__":
    mcp.run(transport='stdio')