package com.agcforge.videodownloader

import android.annotation.SuppressLint
import android.app.Activity
import android.content.Context
import androidx.multidex.MultiDex
import androidx.multidex.MultiDexApplication
import android.app.Application
import android.os.Bundle
import androidx.lifecycle.LifecycleObserver


class App: MultiDexApplication(), Application.ActivityLifecycleCallbacks, LifecycleObserver {

    companion object {
        private const val TAG = "App_Root_Main"

        @SuppressLint("StaticFieldLeak")
        @Volatile
        private var mInstance: App? = null

        fun getInstance(): App {
            return mInstance ?: throw IllegalStateException("App not initialized")
        }

        var click = 3
        var backclick = 3

        var AdsClickCount = 0
        var backAdsClickCount = 0

        var TypeAds = "admob"

        const val Onesignal_ID = "86149c37-f641-4b38-a75a-5f52df523c07"


    }
    override fun onCreate() {
        super.onCreate()
        mInstance = this
    }

    override fun attachBaseContext(base: Context?) {
        super.attachBaseContext(base)
        MultiDex.install(this)
    }

    override fun onActivityCreated(
        activity: Activity,
        savedInstanceState: Bundle?
    ) {
        TODO("Not yet implemented")
    }

    override fun onActivityDestroyed(activity: Activity) {
        TODO("Not yet implemented")
    }

    override fun onActivityPaused(activity: Activity) {
        TODO("Not yet implemented")
    }

    override fun onActivityResumed(activity: Activity) {
        TODO("Not yet implemented")
    }

    override fun onActivitySaveInstanceState(
        activity: Activity,
        outState: Bundle
    ) {
        TODO("Not yet implemented")
    }

    override fun onActivityStarted(activity: Activity) {
    }

    override fun onActivityStopped(activity: Activity) {
        TODO("Not yet implemented")
    }


}