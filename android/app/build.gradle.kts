import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.compose.compiler)
    alias(libs.plugins.google.services)
    alias(libs.plugins.devtools.ksp)
    alias(libs.plugins.kotlin.serialization)
    alias(libs.plugins.kotlin.parcelize)
}

android {
    namespace = "com.agcforge.videodownloader"
    compileSdk {
        version = release(36)
    }

    defaultConfig {
        applicationId = "com.agcforge.videodownloader"
        minSdk = 24
        targetSdk = 36
        versionCode = 1
        versionName = "1.0.0"
        multiDexEnabled = true
        vectorDrawables.useSupportLibrary = true
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        buildConfigField("String", "BASE_URL", "\"https://api-simontok.agcforge.com/api/v1/\"")
        buildConfigField("String", "CENTRIFUGO_URL", "\"https://websocket.infrastructures.help/connection/websocket\"")
        buildConfigField("String", "API_KEY", "\"39eb7a7c6bbd61d93bf15362e28a499ba4f72f3cacad8f326381c0a4674a2270\"")
    }
    ndkVersion = "29.0.14033849 rc4"
    buildTypes {
        getByName("debug") {
            enableUnitTestCoverage = true
        }
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(
                getDefaultProguardFile("proguard-android-optimize.txt"),
                "proguard-rules.pro"
            )
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    buildFeatures {
        compose = true
        dataBinding = true
        viewBinding = true
        buildConfig = true
    }
    lint {
        abortOnError = true
        checkReleaseBuilds = false
        baseline = file("lint-baseline.xml")
    }
    packaging {
        resources {
            excludes += "/META-INF/{AL,AL2.0,LGPL2.1}"
            excludes += "/META-INF/README.md"
            excludes += "/META-INF/LICENSE*"
            excludes += "/META-INF/NOTICE*"
        }
        jniLibs.useLegacyPackaging = true
    }

}
kotlin {
    compilerOptions {
        jvmTarget = JvmTarget.JVM_11
        freeCompilerArgs.add("-XXLanguage:+PropertyParamAnnotationDefaultTargetMode")
        freeCompilerArgs.add("-opt-in=kotlin.RequiresOptIn")
    }
}
dependencies {
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.appcompat)
    implementation(libs.androidx.lifecycle.livedata.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.kotlinx.coroutines.core)
    implementation(libs.kotlinx.coroutines.android)
    implementation(libs.androidx.activity.compose)
    implementation(libs.kotlin.stdlib)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.multidex)
    implementation(libs.work.runtime.ktx)
    implementation(libs.work.rxjava2)
    implementation(libs.work.gcm)
    androidTestImplementation(libs.work.testing)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    // General Library
    implementation(libs.material)
    implementation(libs.androidx.constraintlayout)
    implementation(libs.androidx.navigation.fragment.ktx)
    implementation(libs.androidx.navigation.ui.ktx)
    implementation (libs.cardview)
    implementation (libs.coordinatorlayout)
    implementation (libs.drawerlayout)
    implementation(libs.androidx.fragment.ktx)
    implementation(libs.androidx.fragment.compose)
    debugImplementation(libs.androidx.fragment.testing)
    implementation (libs.gridlayout)
    implementation (libs.preference)
    implementation (libs.preference.ktx)
    implementation (libs.recyclerview)
    implementation (libs.swiperefreshlayout)
    implementation (libs.viewpager2)
    implementation (libs.palette)
    implementation (libs.vectordrawable.animated)
    implementation(libs.androidx.foundation)
    implementation(libs.androidx.material3)
    implementation(libs.androidx.material3.window.size.class1)
    implementation(libs.androidx.material3.adaptive.navigation.suite)
    implementation(libs.kotlinx.datetime)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.swiperefreshlayout)
    implementation(libs.androidx.core.splashscreen)
    // Messaging, Database & Ads
    implementation(libs.play.services.ads)
    implementation(libs.unity.ads)
    implementation(libs.inapp.sdk)
    implementation(libs.billing)
    implementation(libs.okhttp)
    implementation(libs.logging.interceptor)
    implementation(libs.retrofit)
    implementation(libs.converter.gson)
    implementation(libs.gson)
    implementation(libs.glide)
    // Websocket and other
    implementation(libs.centrifuge.java)
//    implementation(libs.yt.dlp.library)
//    implementation(libs.yt.dlp.ffmpeg)
//    implementation(libs.yt.dlp.aria2c)
}