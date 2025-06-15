package com.example.sparkdemo

import org.apache.spark.sql.SparkSession

/**
 * A simple Spark application that counts words in a sequence of strings.
 */
object SparkDemoApp {
  def main(args: Array[String]): Unit = {
    val spark = SparkSession.builder
      .appName("Scala Spark Demo")
      // .master("local[*]") // Remove or comment out when submitting to a cluster
      .getOrCreate()

    println("SparkSession Created Successfully")

    // For implicit conversions like converting RDDs to DataFrames
    import spark.implicits._

    val data = Seq(
      "Hello Spark from Scala",
      "Spark is awesome Scala is fun",
      "Hello World from Spark Scala demo"
    )

    val df = data.toDF("text")

    println("Original DataFrame:")
    df.show(false)

    val wordsDf = df.selectExpr("explode(split(lower(text), '\\s+')) as word")
      .filter($"word" =!= "") // Filter out empty strings that might result from multiple spaces

    println("Words DataFrame:")
    wordsDf.show(false)

    val wordCountsDf = wordsDf.groupBy("word").count()

    println("Word Counts DataFrame:")
    wordCountsDf.show()

    println(s"Pi is roughly ${Math.PI}") // Example of a simple calculation

    // Example: Create a small DataFrame and write it to HDFS (if HDFS is accessible and configured)
    // This part is optional and depends on your HDFS setup within Docker.
    // Ensure your hadoop.env and core-site.xml in the Docker images are correctly pointing to hdfs://namenode:9000
    try {
      val sampleData = Seq(("sample1", 1), ("sample2", 2)).toDF("key", "value")
      // val hdfsOutputPath = "hdfs://namenode:9000/user/spark_demo_output"
      // println(s"Attempting to write sample data to HDFS: ${hdfsOutputPath}")
      // sampleData.write.mode("overwrite").csv(hdfsOutputPath)
      // println(s"Successfully wrote sample data to HDFS: ${hdfsOutputPath}")

      // // To read it back (optional)
      // val readSampleData = spark.read.csv(hdfsOutputPath)
      // println("Data read back from HDFS:")
      // readSampleData.show()
    } catch {
      case e: Exception =>
        println(s"Error interacting with HDFS: ${e.getMessage}")
        println("Continuing without HDFS interaction for this demo.")
    }


    spark.stop()
    println("SparkSession Stopped")
  }
} 